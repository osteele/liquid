package liquid

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"testing"

	"github.com/osteele/liquid/render"
	"github.com/stretchr/testify/require"
)

func Example() {
	engine := NewEngine()
	source := `<h1>{{ page.title }}</h1>`
	bindings := map[string]any{
		"page": map[string]string{
			"title": "Introduction",
		},
	}

	out, err := engine.ParseAndRenderString(source, bindings)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println(out)
	// Output: <h1>Introduction</h1>
}

func ExampleEngine_ParseAndRenderString() {
	engine := NewEngine()
	source := `{{ hello | capitalize | append: " Mundo" }}`
	bindings := map[string]any{"hello": "hola"}

	out, err := engine.ParseAndRenderString(source, bindings)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println(out)
	// Output: Hola Mundo
}

func ExampleEngine_ParseTemplate() {
	engine := NewEngine()
	source := `{{ hello | capitalize | append: " Mundo" }}`
	bindings := map[string]any{"hello": "hola"}

	tpl, err := engine.ParseString(source)
	if err != nil {
		log.Fatalln(err)
	}

	out, err := tpl.RenderString(bindings)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println(out)
	// Output: Hola Mundo
}

func ExampleEngine_RegisterFilter() {
	engine := NewEngine()
	engine.RegisterFilter("has_prefix", strings.HasPrefix)

	template := `{{ title | has_prefix: "Intro" }}`
	bindings := map[string]any{
		"title": "Introduction",
	}

	out, err := engine.ParseAndRenderString(template, bindings)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println(out)
	// Output: true
}

func ExampleEngine_RegisterFilter_optional_argument() {
	engine := NewEngine()
	// func(a, b int) int) would default the second argument to zero.
	// Then we can't tell the difference between {{ n | inc }} and
	// {{ n | inc: 0 }}. A function in the parameter list has a special
	// meaning as a default parameter.
	engine.RegisterFilter("inc", func(a int, b func(int) int) int {
		return a + b(1)
	})

	template := `10 + 1 = {{ m | inc }}; 20 + 5 = {{ n | inc: 5 }}`
	bindings := map[string]any{
		"m": 10,
		"n": "20",
	}

	out, err := engine.ParseAndRenderString(template, bindings)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println(out)
	// Output: 10 + 1 = 11; 20 + 5 = 25
}

func ExampleEngine_RegisterTag() {
	engine := NewEngine()
	engine.RegisterTag("echo", func(c render.Context) (string, error) {
		return c.TagArgs(), nil
	})

	template := `{% echo hello world %}`

	out, err := engine.ParseAndRenderString(template, emptyBindings)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println(out)
	// Output: hello world
}

func ExampleEngine_RegisterBlock() {
	engine := NewEngine()
	engine.RegisterBlock("length", func(c render.Context) (string, error) {
		s, err := c.InnerString()
		if err != nil {
			return "", err
		}

		n := len(s)

		return strconv.Itoa(n), nil
	})

	template := `{% length %}abc{% endlength %}`

	out, err := engine.ParseAndRenderString(template, emptyBindings)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println(out)
	// Output: 3
}

func TestUnicodeInLiquidTemplates(t *testing.T) {
	t.Run("SimpleUnicodeVariable", func(t *testing.T) {
		vars := map[string]any{
			"用户名": "张三",
			"年龄":  25,
		}

		template := NewEngine()
		result, err := template.ParseAndRender([]byte("Hello {{ 用户名 }}, 年龄: {{ 年龄 }}"), vars)

		require.NoError(t, err)
		require.Equal(t, "Hello 张三, 年龄: 25", string(result))
	})

	t.Run("UnicodeWithProperties", func(t *testing.T) {
		vars := map[string]any{
			"用户": map[string]any{
				"姓名": "李四",
				"邮箱": "lisi@example.com",
			},
		}

		template := NewEngine()
		result, err := template.ParseAndRender([]byte("用户: {{ 用户.姓名 }} - {{ 用户.邮箱 }}"), vars)

		require.NoError(t, err)
		require.Equal(t, "用户: 李四 - lisi@example.com", string(result))
	})

	t.Run("UnicodeInConditions", func(t *testing.T) {
		vars := map[string]any{
			"产品": map[string]any{
				"价格":  100,
				"可用?": true,
			},
		}

		template := NewEngine()
		result, err := template.ParseAndRender([]byte("{% if 产品.可用? %}价格: {{ 产品.价格 }}{% endif %}"), vars)

		require.NoError(t, err)
		require.Equal(t, "价格: 100", string(result))
	})

	t.Run("MixedScripts", func(t *testing.T) {
		vars := map[string]any{
			"user_プロファイル": map[string]any{
				"first_نام": "Xip",
				"last_名字":   "Wang",
			},
		}

		template := NewEngine()
		result, err := template.ParseAndRender([]byte("Name: {{ user_プロファイル.first_نام }} {{ user_プロファイル.last_名字 }}"), vars)

		require.NoError(t, err)
		require.Equal(t, "Name: Xip Wang", string(result))
	})

	t.Run("BengaliTemplate", func(t *testing.T) {
		vars := map[string]any{
			"ব্যবহারকারী": "উৎপল",
			"বয়স":        28,
		}

		template := NewEngine()
		result, err := template.ParseAndRender([]byte("নাম: {{ ব্যবহারকারী }}, বয়স: {{ বয়স }}"), vars)

		require.NoError(t, err)
		require.Equal(t, "নাম: উৎপল, বয়স: 28", string(result))
	})

	t.Run("ArabicTemplate", func(t *testing.T) {
		vars := map[string]any{
			"المستخدم": map[string]any{
				"الاسم": "أحمد",
				"العمر": 28,
			},
		}

		template := NewEngine()
		result, err := template.ParseAndRender([]byte("الاسم: {{ المستخدم.الاسم }}, العمر: {{ المستخدم.العمر }}"), vars)

		require.NoError(t, err)
		require.Equal(t, "الاسم: أحمد, العمر: 28", string(result))
	})

	t.Run("ComplexUnicodeExpression", func(t *testing.T) {
		vars := map[string]any{
			"用户": map[string]any{
				"年龄":  20,
				"国家":  "中国",
				"活动?": true,
			},
		}

		template := NewEngine()
		result, err := template.ParseAndRender(
			[]byte("{% if 用户.年龄 >= 18 and 用户.国家 == '中国' and 用户.活动? %}合格用户{% endif %}"),
			vars,
		)

		require.NoError(t, err)
		require.Equal(t, "合格用户", string(result))
	})
}

func TestUnicodeEdgeCases(t *testing.T) {
	t.Run("UnicodeInLoop", func(t *testing.T) {
		vars := map[string]any{
			"产品列表": []map[string]any{
				{"名称": "产品一", "价格": 10},
				{"名称": "产品二", "价格": 20},
			},
		}

		template := NewEngine()
		result, err := template.ParseAndRender(
			[]byte("{% for 产品 in 产品列表 %}{{ 产品.名称 }}: {{ 产品.价格 }}{% endfor %}"),
			vars,
		)

		require.NoError(t, err)
		require.Contains(t, string(result), "产品一: 10")
		require.Contains(t, string(result), "产品二: 20")
	})

	t.Run("UnicodeWithFilters", func(t *testing.T) {
		vars := map[string]any{
			"消息": "hello world",
		}

		template := NewEngine()
		result, err := template.ParseAndRender(
			[]byte("{{ 消息 | upcase }}"),
			vars,
		)

		require.NoError(t, err)
		require.Equal(t, "HELLO WORLD", string(result))
	})
}

// TestIssue63_UnicodeVariableNames tests that Unicode variable names work correctly
// See: https://github.com/osteele/liquid/issues/63
func TestIssue63_UnicodeVariableNames(t *testing.T) {
	t.Run("ExactIssue63Example", func(t *testing.T) {
		// This is the exact example from issue #63 that was failing
		vars := map[string]any{
			"描述": "content",
		}

		template := NewEngine()
		result, err := template.ParseAndRender([]byte("{{ 描述 }}"), vars)

		require.NoError(t, err)
		require.Equal(t, "content", string(result))
	})
}

func TestRemoveTag(t *testing.T) {
	template := NewEngine()
	template.RegisterTag("echo", func(c render.Context) (string, error) {
		return c.TagArgs(), nil
	})

	source := `{% echo hello world %}`

	_, err := template.ParseAndRenderString(source, emptyBindings)
	require.NoError(t, err)

	template.RemoveTag("echo")

	_, err = template.ParseAndRenderString(source, emptyBindings)
	require.Error(t, err)
}
