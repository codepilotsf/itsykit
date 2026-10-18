package site

import (
	"encoding/json"
	stdhtml "html"
	"strings"
	"testing"

	"golang.org/x/net/html"
)

func TestLiteralTemplateBraces(t *testing.T) {
	literals := []string{
		`Use { and } to group items.`, `{}`, `{not a Flea expression}`, `{x | x > 0}`, `{1, 2, {3, 4}}`, `{`, `}`,
		`{rendering header}`, `{snipp missing}`, `{#iff app.site_name}`, `{app.site_name + 1}`,
		`{"value":"{app.site_name}","snip":"{snip a7Km2pQ9xB}","media":"{snip a7Km2pQ9xB.src}","block":"{#multisnip team}{/multisnip}"}`,
		`{"a":[{},[],{"b":[1,2]}],"quote":"escaped \" } { ","slash":"\\"}`,
		`{&quot;value&quot;:&quot;{app.site_name}&quot;}`,
	}
	for _, literal := range literals {
		t.Run(literal, func(t *testing.T) {
			root := fixture(t)
			source := `<div data-value='` + literal + `'>` + literal + `</div><p>{app.site_name}</p>`
			put(t, root, "index.html", source)
			p := mustLoad(t, root)
			compiled, err := p.compile(source, "index.html", nil, map[string]string{})
			if err != nil || !strings.Contains(compiled, `data-value='`+literal+`'`) {
				t.Fatalf("literal source changed: %s %v", compiled, err)
			}
			w := request(NewServer(p), "GET", "/")
			if w.Code != 200 || !strings.Contains(w.Body.String(), "<p>Test</p>") {
				t.Fatalf("%d %s", w.Code, w.Body.String())
			}
		})
	}
}

func TestLiteralOpeningDoesNotConsumeDirective(t *testing.T) {
	for _, source := range []string{`<p>{ {app.site_name}</p>`, `<p>{unclosed {app.site_name}</p>`, `<p>{'unclosed</p><p>{app.site_name}</p>`} {
		root := fixture(t)
		put(t, root, "index.html", source)
		w := request(NewServer(mustLoad(t, root)), "GET", "/")
		if w.Code != 200 || !strings.Contains(w.Body.String(), "Test") {
			t.Fatalf("%s: %d %s", source, w.Code, w.Body.String())
		}
	}
}

func TestLiteralJSONAttributes(t *testing.T) {
	values := []string{
		`[{"id":"spacing","kind":"range","min":0,"max":1,"default":0.5}]`,
		`[{"nested":[{},[],{"value":"{app.site_name}"}],"quote":"\\\"{}","slash":"\\"}]`,
		`{"encoded":"&#123;snip a7Km2pQ9xB&#125;"}`, `{"html":"<p>{app.site_name}</p>"}`, `[]`, `{}`, ` { "a" : [1, null, true] } `,
		`{"rich":"{snip a7Km2pQ9xB}","media":"{snip a7Km2pQ9xB.src}","csrf":"{csrf}"}`,
	}
	for _, value := range values {
		for _, mode := range []string{"single", "double", "entities"} {
			t.Run(mode+value, func(t *testing.T) {
				encoded := value
				quote := "'"
				if mode == "double" {
					quote = `"`
					encoded = strings.ReplaceAll(encoded, `"`, "&quot;")
				}
				if mode == "entities" {
					encoded = strings.NewReplacer("{", "&#123;", "}", "&#125;").Replace(encoded)
				}
				root := fixture(t)
				put(t, root, "index.html", `<div data-impeccable-params=`+quote+encoded+quote+` data-name="{app.site_name}" data-mixed='{} {app.site_name}'></div>`)
				w := request(NewServer(mustLoad(t, root)), "GET", "/")
				if w.Code != 200 {
					t.Fatal(w.Code, w.Body.String())
				}
				z := html.NewTokenizer(strings.NewReader(w.Body.String()))
				for z.Next() != html.ErrorToken {
					tok := z.Token()
					for _, a := range tok.Attr {
						switch a.Key {
						case "data-impeccable-params":
							if a.Val != stdhtml.UnescapeString(value) {
								t.Fatalf("decoded value %q, want %q", a.Val, value)
							}
							var parsed any
							if err := json.Unmarshal([]byte(a.Val), &parsed); err != nil {
								t.Fatal(err)
							}
						case "data-name":
							if a.Val != "Test" {
								t.Fatal(a.Val)
							}
						case "data-mixed":
							if a.Val != "{} Test" {
								t.Fatal(a.Val)
							}
						}
					}
				}
			})
		}
	}
}

func TestRecognizableMalformedBraces(t *testing.T) {
	for _, source := range []string{
		`{#if}`, `{#unless}`, `{#each}`, `{#multisnip}`, `{:else nonsense}`, `{/if extra}`, `{snip}`, `{render}`, `{icon}`, `{slot extra}`, `{csrf extra}`,
		`{#if app.site_name`, `{snip missing`, `{render missing`, `{icon search`, `{app.site_name`,
	} {
		root := fixture(t)
		put(t, root, "index.html", "<main>\n"+source+"\n</main>")
		_, err := Load(root)
		if err == nil || !strings.Contains(err.Error(), "index.html:2") {
			t.Fatalf("%s: %v", source, err)
		}
	}
}

func TestEntitySyntaxAndContentAreNotCompiled(t *testing.T) {
	root := fixture(t)
	put(t, root, "index.html", `<p>&#123;app.site_name&#125; &#123;snip a7Km2pQ9xB&#125;</p><div data-media="&#123;snip a7Km2pQ9xB.src&#125;"></div><p>{snip team}</p>`)
	w := request(NewServer(mustLoad(t, root)), "GET", "/")
	if w.Code != 200 || strings.Contains(w.Body.String(), "Test") || !strings.Contains(w.Body.String(), "{app.site_name}") {
		t.Fatal(w.Code, w.Body.String())
	}
}

func TestMultisnipLiteralBraces(t *testing.T) {
	body := `<article data-json='{"path":"{snip.name}","rich":"{snip a7Km2pQ9xB}"}'><h2>{snip.name}</h2><p>{not Flea}</p><p>&#123;snip.literal&#125;</p></article>`
	def, err := definition(body)
	if err != nil || len(def.Fields) != 1 {
		t.Fatal(def, err)
	}
	blocks, err := multisnipBlocks("index.html", `<p>{"example":"{#multisnip team}{/multisnip}"}</p>`)
	if err != nil || len(blocks) != 0 {
		t.Fatal(blocks, err)
	}
}

func TestBareAliasBraceRecognition(t *testing.T) {
	root := fixture(t)
	put(t, root, "index.html", `{#each snip.team.members as member}<p>{ member }</p>{/each}<p>Use { and } to group items.</p>`)
	p := mustLoad(t, root)
	compiled, err := p.compile(`{#each snip.team.members as member}{ member }{/each}`, "index.html", nil, map[string]string{})
	if err != nil || strings.Contains(compiled, "{ member }") {
		t.Fatal(compiled, err)
	}
	put(t, root, "index.html", `{#each snip.team.members as member}{ member`)
	if _, err := Load(root); err == nil || !strings.Contains(err.Error(), "unclosed template directive") {
		t.Fatal(err)
	}
}
