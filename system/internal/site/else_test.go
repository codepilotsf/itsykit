package site

import (
	"strings"
	"testing"
)

func TestElseBranches(t *testing.T) {
	for _, tc := range []struct{ source, want string }{
		{`{#unless app.missing}yes{:else}no{/unless}`, "yes"},
		{`{#unless app.site_name}no{:else}yes{/unless}`, "yes"},
		{`{#unless app.site_name}no{:else if app.site_name}yes{:else}no{/unless}`, "yes"},
		{`{#unless snip.team.false}a{/unless}{#unless snip.team.zero}b{/unless}{#unless snip.team.blank}c{/unless}{#unless snip.team.empty}d{/unless}`, "abcd"},
		{`{#each snip.team.members as member}{#unless member.children}{member.name}{/unless}{/each}`, "Cy"},
		{`{#unless app.missing}{#if app.site_name}{app.unsafe}{/if}{/unless}`, `&#34;/&gt;&lt;script&gt;alert(1)&lt;/script&gt;`},
		{`{#if app.site_name}yes{:else}no{/if}`, "yes"},
		{`{#if snip.team.false}no{:else if snip.team.zero}no{:else if app.site_name}yes{:else}no{/if}`, "yes"},
		{`{#if snip.missing}no{:else}fallback{/if}`, "fallback"},
		{`{#each snip.team.empty as item}no{:else}{app.site_name}{/each}`, "Test"},
		{`{#each snip.missing as item}no{:else}missing{/each}`, "missing"},
		{`{#each snip.team.members as item}{#each item.children as item}{item.name}{:else}{render fallback}{/each}:{item.name};{/each}`, "Bea:Ada;Cy:Test:Cy;"},
		{`{#if snip.missing}no{:else}<a href="{app.url}">{app.unsafe}</a>{/if}`, `<a href="#ZgotmplZ">&#34;/&gt;&lt;script&gt;alert(1)&lt;/script&gt;</a>`},
	} {
		t.Run(tc.source, func(t *testing.T) {
			root := fixture(t)
			put(t, root, "_fallback.html", `{item.name}:{app.site_name}`)
			put(t, root, "index.html", tc.source)
			p := mustLoad(t, root)
			got, err := p.render(p.Pages["/"], "/")
			if err != nil || string(got) != tc.want {
				t.Fatalf("got %s, want %s: %v", got, tc.want, err)
			}
		})
	}
}

func TestInvalidElseBranches(t *testing.T) {
	for _, source := range []string{
		`{/unless}`, `{#unless app.site_name}{/if}`, `{#if app.site_name}{/unless}`,
		`{#unless }{/unless}`, `{#unless app.site_name == true}{/unless}`,
		`{#unless app.site_name}{:else}{:else}{/unless}`, `{#unless app.site_name}`,
		`{:else}`, `{:else if app.site_name}`, `{#if app.site_name}{:else}{:else}{/if}`,
		`{#if app.site_name}{:else}{:else if app.site_name}{/if}`,
		`{#each snip.team.empty as x}{:else if app.site_name}{/each}`,
		`{#if app.site_name}{:else if }{/if}`, `{#if app.site_name}{:else if app.site_name == true}{/if}`,
		`{#if app.site_name}{:else}{/each}`, `{#if app.site_name}{:else}`,
	} {
		root := fixture(t)
		put(t, root, "index.html", source)
		if _, err := Load(root); err == nil || !strings.Contains(err.Error(), "index.html:") {
			t.Fatalf("%s: %v", source, err)
		}
	}
}
