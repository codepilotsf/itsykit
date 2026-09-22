package site

import (
	"strings"
	"testing"
)

func TestElseBranches(t *testing.T) {
	for _, tc := range []struct{ source, want string }{
		{`{#if site.name}yes{:else}no{/if}`, "yes"},
		{`{#if snip.team.false}no{:else if snip.team.zero}no{:else if site.name}yes{:else}no{/if}`, "yes"},
		{`{#if snip.missing}no{:else}fallback{/if}`, "fallback"},
		{`{#each snip.team.empty as item}no{:else}{site.name}{/each}`, "Test"},
		{`{#each snip.missing as item}no{:else}missing{/each}`, "missing"},
		{`{#each snip.team.members as item}{#each item.children as item}{item.name}{:else}{render fallback}{/each}:{item.name};{/each}`, "Bea:Ada;Cy:Test:Cy;"},
		{`{#if snip.missing}no{:else}<a href="{site.url}">{site.unsafe}</a>{/if}`, `<a href="#ZgotmplZ">&#34;/&gt;&lt;script&gt;alert(1)&lt;/script&gt;</a>`},
	} {
		t.Run(tc.source, func(t *testing.T) {
			root := fixture(t)
			put(t, root, "_fallback.html", `{item.name}:{site.name}`)
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
		`{:else}`, `{:else if site.name}`, `{#if site.name}{:else}{:else}{/if}`,
		`{#if site.name}{:else}{:else if site.name}{/if}`,
		`{#each snip.team.empty as x}{:else if site.name}{/each}`,
		`{#if site.name}{:else if }{/if}`, `{#if site.name}{:else if site.name == true}{/if}`,
		`{#if site.name}{:else}{/each}`, `{#if site.name}{:else}`,
	} {
		root := fixture(t)
		put(t, root, "index.html", source)
		if _, err := Load(root); err == nil || !strings.Contains(err.Error(), "index.html:") {
			t.Fatalf("%s: %v", source, err)
		}
	}
}
