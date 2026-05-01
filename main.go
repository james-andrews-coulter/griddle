package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/beevik/etree"
	"github.com/expr-lang/expr"
	"github.com/mmcdole/gofeed"
)

// --- Data Model ---

type Feed struct {
	Name       string        `json:"name"`
	URL        string        `json:"url"`
	GroupLogic string        `json:"group_logic"`
	Groups     []FilterGroup `json:"groups"`
}

type FilterGroup struct {
	Logic string `json:"logic"`
	Rules []Rule `json:"rules"`
}

type Rule struct {
	Field    string `json:"field"`
	Operator string `json:"operator"`
	Value    string `json:"value"`
}

// --- Persistence ---

var dataFile = "/data/feeds.json"

func loadFeeds() ([]Feed, error) {
	data, err := os.ReadFile(dataFile)
	if os.IsNotExist(err) {
		return []Feed{}, nil
	}
	if err != nil {
		return nil, err
	}
	var feeds []Feed
	return feeds, json.Unmarshal(data, &feeds)
}

func saveFeeds(feeds []Feed) error {
	data, err := json.MarshalIndent(feeds, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(dataFile, data, 0644)
}

// --- Filter Engine ---

func joinByLogic(parts []string, logic string) string {
	switch logic {
	case "any":
		return strings.Join(parts, " || ")
	case "none":
		return "!(" + strings.Join(parts, " || ") + ")"
	default:
		if logic != "all" && logic != "" {
			log.Printf("buildExpr: unknown logic %q, defaulting to all", logic)
		}
		return strings.Join(parts, " && ")
	}
}

func buildExpr(feed Feed) string {
	var groupExprs []string

	for _, group := range feed.Groups {
		if len(group.Rules) == 0 {
			continue
		}

		var ruleExprs []string
		for _, rule := range group.Rules {
			val := strings.ToLower(rule.Value)
			field := rule.Field
			var e string
			switch rule.Operator {
			case "contains":
				e = fmt.Sprintf(`lower(%s) contains "%s"`, field, val)
			case "not_contains":
				e = fmt.Sprintf(`!(lower(%s) contains "%s")`, field, val)
			case "equals":
				e = fmt.Sprintf(`lower(%s) == "%s"`, field, val)
			case "not_equals":
				e = fmt.Sprintf(`lower(%s) != "%s"`, field, val)
			default:
				log.Printf("buildExpr: unknown operator %q, defaulting to contains", rule.Operator)
				e = fmt.Sprintf(`lower(%s) contains "%s"`, field, val)
			}
			ruleExprs = append(ruleExprs, e)
		}

		groupExprs = append(groupExprs, joinByLogic(ruleExprs, group.Logic))
	}

	if len(groupExprs) == 0 {
		return "true"
	}

	if len(groupExprs) == 1 {
		return groupExprs[0]
	}

	wrapped := make([]string, len(groupExprs))
	for i, g := range groupExprs {
		wrapped[i] = "(" + g + ")"
	}

	return joinByLogic(wrapped, feed.GroupLogic)
}

// ruleFields extracts unique field names referenced across all rules in a feed.
func ruleFields(feed Feed) []string {
	seen := make(map[string]bool)
	var fields []string
	for _, group := range feed.Groups {
		for _, rule := range group.Rules {
			if !seen[rule.Field] {
				seen[rule.Field] = true
				fields = append(fields, rule.Field)
			}
		}
	}
	return fields
}

// itemToEnv builds a flat map[string]string from a gofeed.Item for expr evaluation.
// fields lists the keys that must be present (defaulting to "").
func itemToEnv(item *gofeed.Item, fields []string) map[string]string {
	env := make(map[string]string, len(fields))
	for _, f := range fields {
		env[f] = ""
	}

	// Standard fields
	if _, ok := env["title"]; ok {
		env["title"] = item.Title
	}
	if _, ok := env["link"]; ok {
		env["link"] = item.Link
	}
	if _, ok := env["description"]; ok {
		env["description"] = item.Description
	}
	if _, ok := env["content"]; ok {
		env["content"] = item.Content
	}
	if _, ok := env["author"]; ok && item.Author != nil {
		env["author"] = item.Author.Name
	}
	needCats := false
	if _, ok := env["categories"]; ok {
		needCats = true
	}
	if _, ok := env["category"]; ok {
		needCats = true
	}
	if needCats {
		cats := strings.Join(item.Categories, ",")
		env["categories"] = cats
		env["category"] = cats
	}

	// Custom (non-namespaced) XML tags
	for k, v := range item.Custom {
		if _, ok := env[k]; ok {
			env[k] = v
		}
	}

	// Namespaced XML tags (e.g. dc:creator, job_listing:company, media:thumbnail)
	// are flattened to <prefix>_<tag> keys since expr-lang treats ":" as a type op.
	// Repeated tags join by comma; tags whose Value is empty fall back to common
	// URL-bearing attributes (url, href) so media:thumbnail etc. remain filterable.
	for ns, tags := range item.Extensions {
		for tag, exts := range tags {
			key := ns + "_" + tag
			if _, ok := env[key]; !ok {
				continue
			}
			var values []string
			for _, e := range exts {
				switch {
				case e.Value != "":
					values = append(values, e.Value)
				case e.Attrs["url"] != "":
					values = append(values, e.Attrs["url"])
				case e.Attrs["href"] != "":
					values = append(values, e.Attrs["href"])
				}
			}
			env[key] = strings.Join(values, ",")
		}
	}

	return env
}

// filterItems applies the feed's filter rules to items and returns those that match.
// On compile error it fails-open and returns all items.
func filterItems(items []*gofeed.Item, feed Feed) []*gofeed.Item {
	exprStr := buildExpr(feed)

	if exprStr == "true" {
		return items
	}

	fields := ruleFields(feed)

	// Build prototype env (all fields → empty string) for type-checking compilation.
	proto := make(map[string]string, len(fields))
	for _, f := range fields {
		proto[f] = ""
	}

	program, err := expr.Compile(exprStr, expr.Env(proto), expr.AsBool())
	if err != nil {
		log.Printf("filterItems: compile error for feed %q: %v", feed.Name, err)
		return items
	}

	var out []*gofeed.Item
	for _, item := range items {
		env := itemToEnv(item, fields)
		result, err := expr.Run(program, env)
		if err != nil {
			log.Printf("filterItems: eval error: %v", err)
			continue
		}
		if result.(bool) {
			out = append(out, item)
		}
	}
	return out
}

// --- Templates ---

const rulePartial = `<div class="rule">
<label class="sr-only">field</label>
<input type="text" name="field" placeholder="title">
<label class="sr-only">operator</label>
<select name="operator" aria-label="operator">
<option value="contains">contains</option>
<option value="not_contains">not contains</option>
<option value="equals">equals</option>
<option value="not_equals">not equals</option>
</select>
<label class="sr-only">value</label>
<input type="text" name="value" placeholder="value">
<button type="button" class="link-action remove-rule" onclick="this.closest('.rule').remove()" aria-label="remove rule">remove</button>
</div>`

var groupPartial = `<fieldset class="group">
<legend class="sr-only">group</legend>
<div class="form-group group-logic" style="display:none">
<label>rules must match</label>
<select name="group_logic">
<option value="all">all (AND)</option>
<option value="any" selected>any (OR)</option>
<option value="none">none (NOR)</option>
</select>
</div>
<div class="rules">` + rulePartial + `</div>
<div class="rule-actions">
<a href="#" hx-get="/partials/rule" hx-target="previous .rules" hx-swap="beforeend">+ rule</a>
<button type="button" class="link-action remove-group" onclick="this.closest('.group').remove()">remove group</button>
</div>
</fieldset>`

var indexTmpl = template.Must(template.New("index").Funcs(template.FuncMap{
	"ruleCount": func(groups []FilterGroup) int {
		n := 0
		for _, g := range groups {
			n += len(g.Rules)
		}
		return n
	},
}).Parse(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>RSS Griddle</title>
<link rel="stylesheet" href="https://unpkg.com/terminal.css@0.7.5/dist/terminal.min.css">
<style>
body.terminal{--primary-color:#1077c4}
.link-action{background:none;border:0;padding:0;margin:0;font:inherit;color:var(--primary-color);cursor:pointer;text-decoration:none;line-height:inherit}
.link-action:hover,.link-action:focus-visible{background:var(--primary-color);color:var(--invert-font-color);outline:none}
#form-error{color:var(--error-color);padding:var(--global-space) 0;font-weight:600}
.sr-only{position:absolute;width:1px;height:1px;padding:0;margin:-1px;overflow:hidden;clip:rect(0,0,0,0);white-space:nowrap;border:0}
fieldset.group{margin:0 0 var(--global-space) 0;padding:var(--global-space) calc(var(--global-space) * 1.5)}
fieldset.group>legend{padding:0 calc(var(--global-space) / 2)}
.group .form-group.group-logic{margin:0 0 var(--global-space) 0}
.rule{display:grid;grid-template-columns:minmax(7em,1fr) minmax(7em,9em) minmax(7em,1fr) auto;gap:calc(var(--global-space) / 2);align-items:center;margin:0 0 calc(var(--global-space) / 2) 0}
.rule input,.rule select{margin:0}
.rule .remove-rule{justify-self:end;white-space:nowrap}
@media (max-width:34em){.rule{grid-template-columns:1fr 1fr;row-gap:calc(var(--global-space) / 2)}.rule input[name="value"]{grid-column:1 / -1}.rule .remove-rule{grid-column:1 / -1}}
.rule-actions{display:flex;justify-content:space-between;align-items:baseline;gap:var(--global-space);margin-top:calc(var(--global-space) / 2)}
.workspace{display:flex;flex-direction:column;gap:calc(var(--global-space) * 2);margin-bottom:calc(var(--global-space) * 2)}
.workspace>.form-column,.workspace>#preview-pane{min-width:0}
.form-column{display:flex;flex-direction:column;gap:calc(var(--global-space) * 4)}
.preview{border-top:1px solid var(--secondary-color);padding-top:var(--global-space)}
.preview-status{margin:0 0 var(--global-space) 0;font-weight:600}
.preview-status.preview-error{color:var(--error-color)}
.preview-xml{margin:0;font-family:var(--mono-font-stack);font-size:13px;line-height:1.5em;white-space:pre-wrap;word-break:break-word;overflow-wrap:anywhere;padding-bottom:var(--global-space)}
.preview-item{display:block;padding:2px 0}
.preview-item.pass{color:var(--font-color)}
.preview-item.filter{color:var(--secondary-color)}
.preview-more{display:block;padding:var(--global-space) 0;color:var(--secondary-color);font-style:italic}
.preview-loading .preview-xml,.preview-loading .preview-status:not(.preview-error){opacity:.6}
@media (prefers-reduced-motion:no-preference){.preview-xml{transition:opacity 150ms cubic-bezier(0.22,1,0.36,1)}}
@media (min-width:64em){
  html,body{height:100vh;overflow:hidden}
  body.terminal .container{max-width:90em;display:flex;flex-direction:column;height:100vh;box-sizing:border-box}
  body.terminal .container>h1{flex:0 0 auto}
  .workspace{flex:1;min-height:0;flex-direction:row;align-items:stretch;gap:calc(var(--global-space) * 3);margin-bottom:0}
  .workspace>.form-column{flex:0 0 32em;overflow-y:auto;min-height:0;padding-right:var(--global-space)}
  .workspace>#preview-pane{flex:1;overflow-y:auto;min-height:0;padding-bottom:var(--global-space)}
  .preview{border-top:none;padding-top:0}
  .preview-status{position:sticky;top:0;background:var(--background-color);padding:var(--global-space) 0 calc(var(--global-space) / 2);margin:0;z-index:1}
}
</style>
<script src="https://unpkg.com/htmx.org@2.0.4"></script>
</head>
<body class="terminal">
<main class="container">
<h1>RSS Griddle</h1>

<div class="workspace">
<div class="form-column">
<section id="feed-form">
{{template "form" .}}
</section>
<section class="feeds">
<h2>Feeds</h2>
{{if .Feeds}}
<table>
<tbody>
{{range .Feeds}}
<tr>
<td>{{.Name}}</td>
<td style="text-align:right"><a href="{{$.Host}}/api/feed?name={{urlquery .Name}}">url</a> · <a href="#" hx-get="/api/edit?name={{urlquery .Name}}" hx-target="#feed-form" hx-swap="innerHTML">edit ({{ruleCount .Groups}})</a> · <a href="#" hx-post="/api/delete?name={{urlquery .Name}}" hx-target="closest tr" hx-swap="delete" hx-confirm="Delete {{.Name}}?">delete</a></td>
</tr>
{{end}}
</tbody>
</table>
{{else}}
<p>No feeds yet.</p>
{{end}}
</section>
</div>
<aside id="preview-pane">
<div id="preview" class="preview" hidden>
<p id="preview-status" class="preview-status" aria-live="polite"></p>
<pre id="preview-body" class="preview-xml"></pre>
</div>
</aside>
</div>
</main>

<script>
document.addEventListener("submit", function(e) {
  if (e.target.id !== "form") return;
  e.preventDefault();
  var form = e.target;
  var data = {
    name: form.querySelector('[name="name"]').value,
    url: form.querySelector('[name="url"]').value,
    group_logic: form.querySelector('[name="group_logic"]').value,
    groups: []
  };
  form.querySelectorAll('.group').forEach(function(g) {
    var group = {
      logic: g.querySelector('[name="group_logic"]').value,
      rules: []
    };
    g.querySelectorAll('.rule').forEach(function(r) {
      group.rules.push({
        field: r.querySelector('[name="field"]').value,
        operator: r.querySelector('[name="operator"]').value,
        value: r.querySelector('[name="value"]').value
      });
    });
    data.groups.push(group);
  });
  fetch(form.action, {
    method: "POST",
    headers: {"Content-Type": "application/json"},
    body: JSON.stringify(data)
  }).then(function(resp) {
    var err = document.getElementById('form-error');
    if (resp.ok) {
      if (err) { err.hidden = true; err.textContent = ''; }
      location.reload();
    } else {
      resp.text().then(function(t) {
        if (err) { err.textContent = 'Error: ' + (t || resp.status); err.hidden = false; }
      });
    }
  });
});
function updateVisibility() {
  var groups = document.querySelectorAll('#groups .group');
  var wrap = document.getElementById('group-logic-wrap');
  if (wrap) wrap.style.display = groups.length >= 2 ? '' : 'none';
  groups.forEach(function(g) {
    var gl = g.querySelector('.group-logic');
    var rules = g.querySelectorAll('.rule');
    if (gl) gl.style.display = rules.length >= 2 ? '' : 'none';
    g.querySelectorAll('.remove-rule').forEach(function(r) {
      r.style.display = rules.length >= 2 ? '' : 'none';
    });
    var rg = g.querySelector('.remove-group');
    if (rg) rg.style.display = groups.length >= 2 ? '' : 'none';
  });
}
new MutationObserver(updateVisibility).observe(document.body, {childList: true, subtree: true});
document.body.addEventListener('htmx:afterSettle', updateVisibility);
updateVisibility();

// --- Live preview (dryrun) ---
(function () {
  var DEBOUNCE_MS = 500;
  var debounceTimer = null;
  var inFlight = null;

  function getFormData() {
    var form = document.getElementById('form');
    if (!form) return null;
    var urlEl = form.querySelector('[name="url"]');
    var url = urlEl ? urlEl.value.trim() : '';
    if (!url) return null;
    var nameEl = form.querySelector('[name="name"]');
    var groupLogicEl = form.querySelector('#group_logic');
    var data = {
      name: nameEl ? nameEl.value : '',
      url: url,
      group_logic: groupLogicEl ? groupLogicEl.value : 'any',
      groups: []
    };
    form.querySelectorAll('.group').forEach(function (g) {
      var glEl = g.querySelector('[name="group_logic"]');
      var group = { logic: glEl ? glEl.value : 'any', rules: [] };
      g.querySelectorAll('.rule').forEach(function (r) {
        group.rules.push({
          field: r.querySelector('[name="field"]').value,
          operator: r.querySelector('[name="operator"]').value,
          value: r.querySelector('[name="value"]').value
        });
      });
      data.groups.push(group);
    });
    return data;
  }

  function setStatus(text, isError) {
    var el = document.getElementById('preview-status');
    if (!el) return;
    el.textContent = text;
    el.classList.toggle('preview-error', !!isError);
  }

  function renderItems(items, more) {
    var body = document.getElementById('preview-body');
    if (!body) return;
    while (body.firstChild) body.removeChild(body.firstChild);
    items.forEach(function (it, idx) {
      var code = document.createElement('code');
      code.className = 'preview-item ' + (it.passed ? 'pass' : 'filter');
      code.textContent = it.xml + (idx < items.length - 1 ? '\n\n' : '');
      body.appendChild(code);
    });
    if (more && more > 0) {
      var moreEl = document.createElement('span');
      moreEl.className = 'preview-more';
      moreEl.textContent = '+ ' + more + ' more';
      body.appendChild(moreEl);
    }
  }

  function showPreview() {
    var p = document.getElementById('preview');
    if (p) p.hidden = false;
  }
  function hidePreview() {
    var p = document.getElementById('preview');
    if (p) p.hidden = true;
  }

  function runPreview() {
    var preview = document.getElementById('preview');
    if (!preview) return;
    var data = getFormData();
    if (!data) {
      hidePreview();
      return;
    }
    showPreview();
    preview.classList.add('preview-loading');
    setStatus('checking…', false);

    if (inFlight) {
      try { inFlight.abort(); } catch (e) {}
    }
    var ctrl = new AbortController();
    inFlight = ctrl;

    fetch('/api/dryrun', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data),
      signal: ctrl.signal
    })
      .then(function (resp) {
        if (!resp.ok) {
          return resp.text().then(function (t) {
            var err = new Error(t || ('http ' + resp.status));
            err.httpStatus = resp.status;
            throw err;
          });
        }
        return resp.json();
      })
      .then(function (j) {
        preview.classList.remove('preview-loading');
        var total = j.total || 0;
        var passN = j.passN || 0;
        var filterN = j.filterN || 0;
        var status;
        if (total === 0) {
          status = 'feed has no items';
        } else if (filterN === 0) {
          status = total + ' items would all pass, add a rule to filter';
        } else if (passN === 0) {
          status = '0 of ' + total + ' items would pass, every item is filtered';
        } else {
          status = passN + ' of ' + total + ' items would pass';
        }
        setStatus(status, false);
        renderItems(j.items || [], j.more || 0);
      })
      .catch(function (err) {
        if (err && err.name === 'AbortError') return;
        preview.classList.remove('preview-loading');
        var raw = (err && err.message) ? err.message.split('\n')[0] : 'unknown error';
        if (raw.length > 140) raw = raw.slice(0, 140) + '…';
        setStatus(raw, true);
      });
  }

  function scheduleRun() {
    if (debounceTimer) clearTimeout(debounceTimer);
    debounceTimer = setTimeout(runPreview, DEBOUNCE_MS);
  }

  document.addEventListener('input', function (e) {
    if (!e.target.closest || !e.target.closest('#form')) return;
    scheduleRun();
  });
  document.addEventListener('change', function (e) {
    if (!e.target.closest || !e.target.closest('#form')) return;
    scheduleRun();
  });
  document.body.addEventListener('htmx:afterSettle', scheduleRun);
  // Initial preview if URL is already populated (edit mode).
  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', function () { setTimeout(runPreview, 200); });
  } else {
    setTimeout(runPreview, 200);
  }
})();
</script>
</body>
</html>

{{define "form"}}
<form id="form" action="{{if .Edit}}/api/save?name={{urlquery .Edit.Name}}{{else}}/feeds{{end}}" method="post">
<fieldset>
<legend>{{if .Edit}}Edit Feed{{else}}New Feed{{end}}</legend>
<div class="form-group">
<label for="name">Name</label>
<input id="name" type="text" name="name" {{if .Edit}}value="{{.Edit.Name}}" readonly{{else}}required placeholder="my-feed"{{end}}>
</div>
<div class="form-group">
<label for="url">URL</label>
<input id="url" type="text" name="url" {{if .Edit}}value="{{.Edit.URL}}"{{else}}placeholder="https://example.com/feed.xml"{{end}} required>
</div>
<div class="form-group" id="group-logic-wrap"{{if .Edit}}{{if lt (len .Edit.Groups) 2}} style="display:none"{{end}}{{else}} style="display:none"{{end}}>
<label for="group_logic">Groups must match</label>
<select id="group_logic" name="group_logic">
{{if .Edit}}<option value="all"{{if eq .Edit.GroupLogic "all"}} selected{{end}}>all</option>
<option value="any"{{if eq .Edit.GroupLogic "any"}} selected{{end}}>any</option>
<option value="none"{{if eq .Edit.GroupLogic "none"}} selected{{end}}>none</option>
{{else}}<option value="all">all</option>
<option value="any" selected>any</option>
<option value="none">none</option>
{{end}}</select>
</div>
<div id="groups">
{{if .Edit}}{{range $gi, $g := .Edit.Groups}}<fieldset class="group">
<legend class="sr-only">group</legend>
<div class="form-group group-logic"{{if lt (len $g.Rules) 2}} style="display:none"{{end}}>
<label>rules must match</label>
<select name="group_logic">
<option value="all"{{if eq $g.Logic "all"}} selected{{end}}>all (AND)</option>
<option value="any"{{if eq $g.Logic "any"}} selected{{end}}>any (OR)</option>
<option value="none"{{if eq $g.Logic "none"}} selected{{end}}>none (NOR)</option>
</select>
</div>
<div class="rules">
{{range $g.Rules}}<div class="rule">
<label class="sr-only">field</label>
<input type="text" name="field" value="{{.Field}}" placeholder="title">
<label class="sr-only">operator</label>
<select name="operator" aria-label="operator">
<option value="contains"{{if eq .Operator "contains"}} selected{{end}}>contains</option>
<option value="not_contains"{{if eq .Operator "not_contains"}} selected{{end}}>not contains</option>
<option value="equals"{{if eq .Operator "equals"}} selected{{end}}>equals</option>
<option value="not_equals"{{if eq .Operator "not_equals"}} selected{{end}}>not equals</option>
</select>
<label class="sr-only">value</label>
<input type="text" name="value" value="{{.Value}}" placeholder="value">
<button type="button" class="link-action remove-rule" onclick="this.closest('.rule').remove()" aria-label="remove rule">remove</button>
</div>{{end}}
</div>
<div class="rule-actions">
<a href="#" hx-get="/partials/rule" hx-target="previous .rules" hx-swap="beforeend">+ rule</a>
<button type="button" class="link-action remove-group" onclick="this.closest('.group').remove()">remove group</button>
</div>
</fieldset>{{end}}{{else}}` + groupPartial + `{{end}}
</div>
<div class="form-group">
<a href="#" hx-get="/partials/group" hx-target="#groups" hx-swap="beforeend">+ group</a>
</div>
{{if .Edit}}<hr>{{end}}
<div id="form-error" role="alert" aria-live="polite" hidden></div>
<div class="form-group">
<button type="submit" class="btn btn-default{{if not .Edit}} btn-block{{end}}">{{if .Edit}}Save{{else}}Create Feed{{end}}</button>
</div>
{{if .Edit}}<button type="button" class="link-action" onclick="location.reload()">cancel</button>{{end}}
</fieldset>
</form>
{{end}}`))

// --- HTTP Handlers ---

// findFeed loads feeds and returns the index of the named feed (-1 if not found).
func findFeed(name string) ([]Feed, int, error) {
	feeds, err := loadFeeds()
	if err != nil {
		return nil, -1, err
	}
	for i, f := range feeds {
		if f.Name == name {
			return feeds, i, nil
		}
	}
	return feeds, -1, nil
}

// feedName extracts the feed name from query param (?name=) or path param ({name}).
func feedName(r *http.Request) string {
	if n := r.URL.Query().Get("name"); n != "" {
		return n
	}
	return r.PathValue("name")
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("ok"))
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	feeds, err := loadFeeds()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	host := "http://" + r.Host
	if err := indexTmpl.Execute(w, map[string]any{"Feeds": feeds, "Host": host, "Edit": nil}); err != nil {
		log.Printf("handleIndex: template error: %v", err)
	}
}

// maxBodySize limits JSON request bodies to 1 MB.
const maxBodySize = 1 << 20

func handleCreate(w http.ResponseWriter, r *http.Request) {
	var feed Feed
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxBodySize)).Decode(&feed); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if feed.Name == "" || feed.URL == "" {
		http.Error(w, "name and url are required", http.StatusBadRequest)
		return
	}
	feeds, idx, err := findFeed(feed.Name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if idx >= 0 {
		http.Error(w, "feed already exists", http.StatusConflict)
		return
	}
	feeds = append(feeds, feed)
	if err := saveFeeds(feeds); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func handleUpdate(w http.ResponseWriter, r *http.Request) {
	name := feedName(r)
	var feed Feed
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxBodySize)).Decode(&feed); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if feed.URL == "" {
		http.Error(w, "url is required", http.StatusBadRequest)
		return
	}
	feeds, idx, err := findFeed(name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if idx < 0 {
		http.Error(w, "feed not found", http.StatusNotFound)
		return
	}
	feed.Name = name
	feeds[idx] = feed
	if err := saveFeeds(feeds); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func handleDelete(w http.ResponseWriter, r *http.Request) {
	feeds, idx, err := findFeed(feedName(r))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if idx < 0 {
		http.Error(w, "feed not found", http.StatusNotFound)
		return
	}
	feeds = append(feeds[:idx], feeds[idx+1:]...)
	if err := saveFeeds(feeds); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func handleEdit(w http.ResponseWriter, r *http.Request) {
	feeds, idx, err := findFeed(feedName(r))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if idx < 0 {
		http.Error(w, "feed not found", http.StatusNotFound)
		return
	}
	if err := indexTmpl.ExecuteTemplate(w, "form", map[string]any{"Edit": &feeds[idx]}); err != nil {
		log.Printf("handleEdit: template error: %v", err)
	}
}

var (
	feedParser = gofeed.NewParser()
	feedClient = &http.Client{Timeout: 15 * time.Second}
)

func handleFeedXML(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimSuffix(feedName(r), ".xml")
	feeds, idx, err := findFeed(name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if idx < 0 {
		http.Error(w, "feed not found", http.StatusNotFound)
		return
	}

	resp, err := feedClient.Get(feeds[idx].URL)
	if err != nil {
		http.Error(w, "failed to fetch upstream feed: "+err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()
	rawXML, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "failed to read upstream feed: "+err.Error(), http.StatusBadGateway)
		return
	}

	parsed, err := feedParser.ParseString(string(rawXML))
	if err != nil {
		http.Error(w, "failed to parse upstream feed: "+err.Error(), http.StatusBadGateway)
		return
	}

	keep := make(map[*gofeed.Item]bool, len(parsed.Items))
	for _, item := range filterItems(parsed.Items, feeds[idx]) {
		keep[item] = true
	}

	doc := etree.NewDocument()
	if err := doc.ReadFromBytes(rawXML); err != nil {
		http.Error(w, "failed to parse upstream XML: "+err.Error(), http.StatusBadGateway)
		return
	}

	// Items are <item> in RSS/RDF, <entry> in Atom.
	itemTag := "item"
	if parsed.FeedType == "atom" {
		itemTag = "entry"
	}
	xmlItems := doc.FindElements("//" + itemTag)
	for i, el := range xmlItems {
		if i >= len(parsed.Items) || !keep[parsed.Items[i]] {
			el.Parent().RemoveChild(el)
		}
	}

	out, err := doc.WriteToBytes()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if ct := resp.Header.Get("Content-Type"); ct != "" {
		w.Header().Set("Content-Type", ct)
	} else {
		w.Header().Set("Content-Type", "application/rss+xml; charset=utf-8")
	}
	w.Write(out)
}

// --- Dryrun (live preview) ---

type dryrunCacheEntry struct {
	items    []*gofeed.Item
	raw      []byte
	feedType string
	expires  time.Time
}

var (
	dryrunCacheMu sync.Mutex
	dryrunCache   = map[string]dryrunCacheEntry{}
)

const (
	dryrunCacheTTL  = 5 * time.Minute
	dryrunItemLimit = 50
)

func fetchAndCacheItems(url string) ([]*gofeed.Item, []byte, string, error) {
	dryrunCacheMu.Lock()
	if e, ok := dryrunCache[url]; ok && time.Now().Before(e.expires) {
		items, raw, ft := e.items, e.raw, e.feedType
		dryrunCacheMu.Unlock()
		return items, raw, ft, nil
	}
	dryrunCacheMu.Unlock()

	resp, err := feedClient.Get(url)
	if err != nil {
		return nil, nil, "", err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, "", err
	}
	parsed, err := feedParser.ParseString(string(raw))
	if err != nil {
		return nil, nil, "", err
	}

	dryrunCacheMu.Lock()
	now := time.Now()
	for k, e := range dryrunCache {
		if now.After(e.expires) {
			delete(dryrunCache, k)
		}
	}
	dryrunCache[url] = dryrunCacheEntry{
		items:    parsed.Items,
		raw:      raw,
		feedType: parsed.FeedType,
		expires:  now.Add(dryrunCacheTTL),
	}
	dryrunCacheMu.Unlock()
	return parsed.Items, raw, parsed.FeedType, nil
}

type dryrunItem struct {
	Passed bool   `json:"passed"`
	XML    string `json:"xml"`
}

type dryrunResponse struct {
	Total   int          `json:"total"`
	PassN   int          `json:"passN"`
	FilterN int          `json:"filterN"`
	Items   []dryrunItem `json:"items"`
	More    int          `json:"more"`
}

func handleDryrun(w http.ResponseWriter, r *http.Request) {
	var feed Feed
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxBodySize)).Decode(&feed); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if feed.URL == "" {
		http.Error(w, "url is required", http.StatusBadRequest)
		return
	}

	items, raw, feedType, err := fetchAndCacheItems(feed.URL)
	if err != nil {
		http.Error(w, "fetch failed: "+err.Error(), http.StatusBadGateway)
		return
	}

	passed := filterItems(items, feed)
	keep := make(map[*gofeed.Item]bool, len(passed))
	for _, it := range passed {
		keep[it] = true
	}

	doc := etree.NewDocument()
	if err := doc.ReadFromBytes(raw); err != nil {
		http.Error(w, "parse failed: "+err.Error(), http.StatusBadGateway)
		return
	}
	itemTag := "item"
	if feedType == "atom" {
		itemTag = "entry"
	}
	xmlItems := doc.FindElements("//" + itemTag)

	out := dryrunResponse{Total: len(items), Items: []dryrunItem{}}
	for i, it := range items {
		passedFlag := keep[it]
		if passedFlag {
			out.PassN++
		} else {
			out.FilterN++
		}
		if len(out.Items) >= dryrunItemLimit || i >= len(xmlItems) {
			continue
		}
		xmlDoc := etree.NewDocument()
		xmlDoc.SetRoot(xmlItems[i].Copy())
		xmlDoc.Indent(2)
		xmlStr, err := xmlDoc.WriteToString()
		if err != nil {
			continue
		}
		out.Items = append(out.Items, dryrunItem{Passed: passedFlag, XML: strings.TrimRight(xmlStr, "\n")})
	}
	out.More = len(items) - len(out.Items)
	if out.More < 0 {
		out.More = 0
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(out)
}

func handlePartialGroup(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(groupPartial))
}

func handlePartialRule(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(rulePartial))
}

// --- Main ---

func main() {
	if v := os.Getenv("DATA_FILE"); v != "" {
		dataFile = v
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", handleIndex)
	mux.HandleFunc("POST /feeds", handleCreate)
	mux.HandleFunc("POST /api/save", handleUpdate)
	mux.HandleFunc("POST /api/delete", handleDelete)
	mux.HandleFunc("GET /api/edit", handleEdit)
	mux.HandleFunc("GET /api/feed", handleFeedXML)
	mux.HandleFunc("GET /feeds/{name}", handleFeedXML)
	mux.HandleFunc("GET /health", handleHealth)
	mux.HandleFunc("POST /api/dryrun", handleDryrun)
	mux.HandleFunc("GET /partials/group", handlePartialGroup)
	mux.HandleFunc("GET /partials/rule", handlePartialRule)
	addr := ":4080"
	if p := os.Getenv("PORT"); p != "" {
		addr = ":" + p
	}
	log.Println("rss-griddle listening on " + addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
