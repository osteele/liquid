package liquid

import (
	"strings"
	"testing"
)

// extractVars mirrors the JS extractVars function from liquid-poc.html.
// It uses GlobalVariableSegments (equivalent to globalVariableSegmentsSync) and
// joins each path with ".".
// NOTE: In Go, ALL segments are already strings ([][]string), so no need to stop at
// numeric/array segments — the tracking layer already handles that via IndexValue.
func extractVars(eng *Engine, templateStr string) ([]string, error) {
	tpl, srcErr := eng.ParseString(templateStr)
	if srcErr != nil {
		return nil, srcErr
	}
	segs, err := tpl.GlobalVariableSegments()
	if err != nil {
		return nil, err
	}
	seen := map[string]bool{}
	var result []string
	for _, path := range segs {
		key := strings.Join(path, ".")
		if key != "" && !seen[key] {
			seen[key] = true
			result = append(result, key)
		}
	}
	return result, nil
}

func mustHave(t *testing.T, got []string, vars ...string) {
	t.Helper()
	set := map[string]bool{}
	for _, v := range got {
		set[v] = true
	}
	for _, v := range vars {
		if !set[v] {
			t.Errorf("missing %q in %v", v, got)
		}
	}
}

func mustNotHave(t *testing.T, got []string, vars ...string) {
	t.Helper()
	set := map[string]bool{}
	for _, v := range got {
		set[v] = true
	}
	for _, v := range vars {
		if set[v] {
			t.Errorf("should not have %q in %v", v, got)
		}
	}
}

func TestJSPoC_Parity(t *testing.T) {
	eng := NewEngine()

	run := func(name, tpl string, have []string, notHave []string) {
		t.Helper()
		t.Run(name, func(t *testing.T) {
			got, err := extractVars(eng, tpl)
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}
			t.Logf("got: %v", got)
			if len(have) > 0 {
				mustHave(t, got, have...)
			}
			if len(notHave) > 0 {
				mustNotHave(t, got, notHave...)
			}
		})
	}

	// ── 01 Básico ────────────────────────────────────────────────────────────────
	run("01 variável simples", `{{ nome }}`, ss("nome"), nil)
	run("01 propriedade de objeto", `{{ customer.email }}`, ss("customer.email"), nil)
	run("01 cadeia longa de propriedades", `{{ order.shipping.address.city }}`, ss("order.shipping.address.city"), nil)
	run("01 múltiplas variáveis", `{{ customer.first_name }} comprou {{ product.name }} por {{ order.total }}`,
		ss("customer.first_name", "product.name", "order.total"), nil)
	run("01 acesso por índice numérico", `{{ items[0] }}`,
		ss("items"), ss("items[0]"))

	// ── 02 Whitespace ────────────────────────────────────────────────────────────
	run("02 sem espaços", `{{customer.email}}`, ss("customer.email"), nil)
	run("02 muitos espaços", `{{     customer.email     }}`, ss("customer.email"), nil)
	run("02 tab", "{{\tcustomer.email\t}}", ss("customer.email"), nil)
	run("02 quebra de linha", "{{\n  customer.email\n}}", ss("customer.email"), nil)
	run("02 whitespace control", `{{- customer.email -}}`, ss("customer.email"), nil)
	run("02 só traço de abertura", `{{- customer.email }}`, ss("customer.email"), nil)
	run("02 só traço de fechamento", `{{ customer.email -}}`, ss("customer.email"), nil)

	// ── 03 Filters ───────────────────────────────────────────────────────────────
	run("03 filter simples", `{{ customer.name | upcase }}`, ss("customer.name"), nil)
	run("03 vários filters encadeados", `{{ customer.bio | strip_html | truncate: 100 | upcase }}`, ss("customer.bio"), nil)
	run("03 filter com argumento variável", `{{ customer.name | default: fallback.name }}`,
		ss("customer.name", "fallback.name"), nil)
	run("03 filter append com variável", `{{ product.slug | append: site.base_url }}`,
		ss("product.slug", "site.base_url"), nil)
	run("03 filter date", `{{ order.created_at | date: "%d/%m/%Y" }}`, ss("order.created_at"), nil)

	// ── 04 if / unless / case ────────────────────────────────────────────────────
	run("04 if simples", `{% if customer.active %}ativo{% endif %}`, ss("customer.active"), nil)
	run("04 if com ==", `{% if customer.status == "premium" %}ok{% endif %}`, ss("customer.status"), nil)
	run("04 if com espaços absurdos >=", `{% if    customer.age   >=   18   %}maior{% endif %}`, ss("customer.age"), nil)
	run("04 if/elsif/else", `{% if customer.tier == "gold" %}gold{% elsif customer.tier == "silver" %}silver{% else %}free{% endif %}`,
		ss("customer.tier"), nil)
	run("04 unless", `{% unless customer.blocked %}mostrar{% endunless %}`, ss("customer.blocked"), nil)
	run("04 if contains", `{% if customer.tags contains "vip" %}VIP{% endif %}`, ss("customer.tags"), nil)
	run("04 if and/or", `{% if customer.active and customer.verified %}ok{% endif %}`,
		ss("customer.active", "customer.verified"), nil)
	run("04 if aninhado", `{% if order.exists %}{% if order.paid %}pago{% endif %}{% endif %}`,
		ss("order.exists", "order.paid"), nil)
	run("04 case/when", `{% case customer.plan %}{% when "free" %}grátis{% when "pro" %}pago{% endcase %}`,
		ss("customer.plan"), nil)

	// ── 05 for ───────────────────────────────────────────────────────────────────
	run("05 for simples — só coleção, props do item são locais",
		`{% for item in cart.items %}{{ item.name }}{% endfor %}`,
		ss("cart.items"), ss("item", "item.name"))
	run("05 for com limit/offset externo",
		`{% for p in products limit: 5 offset: page.offset %}{{ p.title }}{% endfor %}`,
		ss("products", "page.offset"), ss("p", "p.title"))
	run("05 for com forloop.index",
		`{% for item in order.items %}{{ forloop.index }}: {{ item.name }}{% endfor %}`,
		ss("order.items"), ss("item", "item.name", "forloop", "forloop.index"))
	run("05 for aninhado",
		`{% for order in customer.orders %}{% for item in order.items %}{{ item.sku }}{% endfor %}{% endfor %}`,
		ss("customer.orders"), ss("order", "item", "order.items", "item.sku"))
	run("05 for com break condicional",
		`{% for item in list %}{% if item.stop %}{% break %}{% endif %}{{ item.value }}{% endfor %}`,
		ss("list"), ss("item", "item.stop", "item.value"))

	// ── 06 assign / capture ──────────────────────────────────────────────────────
	run("06 assign — fonte é global, variável é local",
		`{% assign full_name = customer.first_name %}{{ full_name }}`,
		ss("customer.first_name"), ss("full_name"))
	run("06 assign com filter",
		`{% assign slug = product.title | downcase | replace: ' ', '-' %}{{ slug }}`,
		ss("product.title"), ss("slug"))
	run("06 capture — variável capturada é local",
		`{% capture greeting %}Olá, {{ customer.name }}{% endcapture %}{{ greeting }}`,
		ss("customer.name"), ss("greeting"))

	// ── 07 Sintaxes disruptivas ──────────────────────────────────────────────────
	run("07 tag if sem espaço após nome",
		`{%if customer.active%}ok{%endif%}`,
		ss("customer.active"), nil)
	run("07 tag com whitespace control e sem espaços",
		`{%-if customer.active-%}ok{%-endif-%}`,
		ss("customer.active"), nil)
	run("07 quebra de linha dentro de tag",
		"{%\n  if\n  customer.active\n%}ok{%\n  endif\n%}",
		ss("customer.active"), nil)
	run("07 várias variáveis em linha única",
		`Olá {{customer.first_name}}, seu pedido #{{order.id}} de {{order.total}} chegará em {{order.eta}}.`,
		ss("customer.first_name", "order.id", "order.total", "order.eta"), nil)
	run("07 template em linha com for+if",
		`{% for i in order.items %}{% if i.available %}{{i.name}} - {{i.price}}{% endif %}{% endfor %}`,
		ss("order.items"), ss("i", "i.name", "i.price", "i.available"))
	run("07 acesso dinâmico por variável como índice",
		`{{ matrix[row.index][col.key] }}`,
		ss("matrix", "row.index", "col.key"), ss("matrix[row.index][col.key]"))
	run("07 filter com múltiplos argumentos variáveis",
		`{{ msg.body | replace: search.term, replace.value }}`,
		ss("msg.body", "search.term", "replace.value"), nil)
	run("07 comparação variável em ambos os lados",
		`{% if user.role == config.required_role %}ok{% endif %}`,
		ss("user.role", "config.required_role"), nil)
	run("07 unless com != sem espaços",
		`{%unless   order.status!="cancelado"  %}ativo{%endunless%}`,
		ss("order.status"), nil)
	run("07 acesso por string literal como chave",
		`{{ customer["first_name"] }}`,
		ss("customer.first_name"), nil)

	// ── 08 Edge cases ────────────────────────────────────────────────────────────
	run("08 assign+for com mesmo nome",
		`{% assign item = global.item %}{% for item in list %}{{ item.x }}{% endfor %}{{ item }}`,
		ss("global.item", "list"), ss("item", "item.x"))
	run("08 template sem variáveis",
		`<p>Texto fixo sem variáveis nenhuma.</p>`,
		nil, nil)
	run("08 variável só em assign",
		`{% assign x = hidden.value %}nada aqui`,
		ss("hidden.value"), ss("x"))
	run("08 profundidade extrema",
		`{{ a.b.c.d.e.f.g.h }}`,
		ss("a.b.c.d.e.f.g.h"), nil)
	run("08 variável em elsif que nunca executa",
		`{% if false %}{% elsif rarely.used.var %}ok{% endif %}`,
		ss("rarely.used.var"), nil)
}

func ss(vals ...string) []string { return vals }
