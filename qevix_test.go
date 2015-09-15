package qevix_test

import (
	"net/url"
	"qevix"
	"regexp"
	"testing"
)

var qvx = qevix.New()

//Конфигурация
func TestConfig(t *testing.T) {
	// Список разрешенных тегов
	qvx.CfgAllowTags([]string{"b", "i", "u", "a", "img", "ul", "ol", "li", "br", "code", "pre", "div", "cut"})

	// Теги которые нужно считать короткими (<br>, <img>)
	qvx.CfgSetTagShort([]string{"br", "img", "cut"})

	// Преформатированные теги, в которых нужно всё заменять на HTML сущности
	qvx.CfgSetTagPreformatted([]string{"code"})

	// Не короткие теги, которые могут быть пустыми и их не нужно из-за этого удалять
	qvx.CfgSetTagIsEmpty([]string{"div"})

	// Теги внутри которых не нужна авто расстановка тегов перевода на новую строку
	qvx.CfgSetTagNoAutoBr([]string{"ul", "ol"})

	// Теги, которые необходимо вырезать вместе с содержимым
	qvx.CfgSetTagCutWithContent([]string{"script", "object", "iframe", "style"})

	// Теги, после которых не нужно добавлять дополнительный перевод строки, например, блочные теги
	qvx.CfgSetTagBlockType([]string{"ol", "ul", "code"})

	// Разрешенные параметры для тегов.
	qvx.CfgAllowTagParams("a", []string{"href", "title", "target", "rel"})
	qvx.CfgAllowTagParams("img", []string{"src", "alt", "title", "align", "width", "height"})

	// Обязательные параметры для тега
	qvx.CfgSetTagParamsRequired("a", []string{"href"})
	qvx.CfgSetTagParamsRequired("img", []string{"src"})

	// Уточнение значений для параметров тега.
	// Разрешенные шаблоны #str, #int, #link, #regexp(...).
	// По умолчанию значения #str (любая строка)
	qvx.CfgAllowTagParamValue("a", "href", "#link")
	qvx.CfgAllowTagParamValue("a", "target", "_blank")

	qvx.CfgAllowTagParamValue("img", "align", []string{"right", "left", "center"})
	qvx.CfgAllowTagParamValue("img", "width", "#int")
	qvx.CfgAllowTagParamValue("img", "height", "#int")

	// Атрибуты тегов, которые будут добавлятся автоматически
	qvx.CfgSetTagParamDefault("a", "rel", "nofollow")
	qvx.CfgSetTagParamDefault("img", "alt", "")

	// Значения параметров тега, которые должны быть обязательно
	qvx.CfgSetTagParamReview("a", "rel", "nofollow")

	// Теги, которые являются контейнерами для указанных тегов
	qvx.CfgSetTagChilds("ul", []string{"li"})
	qvx.CfgSetTagChilds("ol", []string{"li"})

	// Теги, которые могут быть только контейнерами для других тегов
	qvx.CfgSetTagParentOnly([]string{"ul", "ol"})

	// Теги, которые могут быть только дочерними для других тегов
	qvx.CfgSetTagChildOnly([]string{"li"})

	// Теги, которые не должны быть дочерними к другим тегам
	qvx.CfgSetTagGlobal([]string{"cut"})

	// Теги, в которых нужно отключить типографирование текста
	qvx.CfgSetTagNoTypography([]string{"code", "pre"})

	// Список разрешенных протоколов для ссылок (https, http, ftp)
	qvx.CfgSetLinkProtocolAllow([]string{"http", "https"})

	// Выключение режима XHTML
	qvx.CfgSetXHTMLMode(false)

	// Включение режима автозамены символов переводов строк на тег <br>
	qvx.CfgSetAutoBrMode(true)

	// Включение режима автоматического определения ссылок
	qvx.CfgSetAutoLinkMode(true)

	// callback-функция на тег
	qvx.CfgSetTagBuildCallback("code", TagCodeBuild)

	// callback-функция на спецсимволы (@|#|$)
	qvx.CfgSetSpecialCharCallback('#', TagSharpBuild)
	qvx.CfgSetSpecialCharCallback('@', TagAtBuild)
}

func TagCodeBuild(tag string, params map[string]string, content string) string {
	return "<pre><code>" + content + "<code><pre>\n"
}

func TagSharpBuild(str string) string {
	if matched, _ := regexp.MatchString(`^(?i)[\d\p{L}\_\-]{1,32}$`, str); !matched {
		return ""
	}
	return "<a href=\"/tags/" + url.QueryEscape(str) + "/\">#" + str + "</a>"
}

func TagAtBuild(str string) string {
	if matched, _ := regexp.MatchString(`^(?i)[\d\p{L}\_\-]{1,32}$`, str); !matched {
		return ""
	}
	return "<a href=\"/user/" + url.QueryEscape(str) + "/\">@" + str + "</a>"
}

//Тесты

func TestParseN1(t *testing.T) {
	text := `<b>текст текст текст</b>`

	result, _ := qvx.Parse(text)

	expect := `<b>текст текст текст</b>`

	if result != expect {
		t.Errorf("Expect result to equal in func TestParseN1(t *testing.T).\n%s", result)
	}
}

func TestParseN2(t *testing.T) {
	text := `<b>текст <b>текст</b> текст</b>`

	result, _ := qvx.Parse(text)

	expect := `<b>текст <b>текст</b> текст</b>`

	if result != expect {
		t.Errorf("Expect result to equal in func TestParseN2(t *testing.T).\n%s", result)
	}
}

func TestParseN3(t *testing.T) {
	text := `<b>текст <u>текст текст`

	result, _ := qvx.Parse(text)

	expect := `<b>текст <u>текст текст</u></b>`

	if result != expect {
		t.Errorf("Expect result to equal in func TestParseN3(t *testing.T).\n%s", result)
	}
}

func TestParseN4(t *testing.T) {
	text := `<u>текст <s>текст</s> текст</u>`

	result, _ := qvx.Parse(text)

	expect := `<u>текст текст текст</u>`

	if result != expect {
		t.Errorf("Expect result to equal in func TestParseN4(t *testing.T).\n%s", result)
	}
}

func TestParseN5(t *testing.T) {
	text := `текст <script>текст</script> текст`

	result, _ := qvx.Parse(text)

	expect := `текст текст`

	if result != expect {
		t.Errorf("Expect result to equal in func TestParseN5(t *testing.T).\n%s", result)
	}
}

func TestParseN6(t *testing.T) {
	text := `<code>текст <script>текст</script> текст</code>`

	result, _ := qvx.Parse(text)

	expect := `<pre><code>текст &#60;script&#62;текст&#60;/script&#62; текст<code><pre>`

	if result != expect {
		t.Errorf("Expect result to equal in func TestParseN6(t *testing.T).\n%s", result)
	}
}

func TestParseN7(t *testing.T) {
	text := `текст <div></div> <b></b> текст</b>`

	result, _ := qvx.Parse(text)

	expect := `текст <div></div> текст`

	if result != expect {
		t.Errorf("Expect result to equal in func TestParseN7(t *testing.T).\n%s", result)
	}
}

func TestParseN8(t *testing.T) {
	text := `текст http://dighub.ru текст`

	result, _ := qvx.Parse(text)

	expect := `текст <a href="http://dighub.ru" rel="nofollow">http://dighub.ru</a> текст`

	if result != expect {
		t.Errorf("Expect result to equal in func TestParseN8(t *testing.T).\n%s", result)
	}
}

func TestParseN9(t *testing.T) {
	text := `текст <b>http://dighub.ru</b> текст`

	result, _ := qvx.Parse(text)

	expect := `текст <b><a href="http://dighub.ru" rel="nofollow">http://dighub.ru</a></b> текст`

	if result != expect {
		t.Errorf("Expect result to equal in func TestParseN9(t *testing.T).\n%s", result)
	}
}

func TestParseN10(t *testing.T) {
	text := `текст <a href="http://dighub.ru">http://dighub.ru</a> текст`

	result, _ := qvx.Parse(text)

	expect := `текст <a href="http://dighub.ru" rel="nofollow">http://dighub.ru</a> текст`

	if result != expect {
		t.Errorf("Expect result to equal in func TestParseN10(t *testing.T).\n%s", result)
	}
}

func TestParseN11(t *testing.T) {
	text := `текст http://yandex.ru/search/?lr=2&text=golang!..`

	result, _ := qvx.Parse(text)

	expect := `текст <a href="http://yandex.ru/search/?lr=2&text=golang" rel="nofollow">http://yandex.ru/search/?lr=2&text=golang</a>!..`

	if result != expect {
		t.Errorf("Expect result to equal in func TestParseN11(t *testing.T).\n%s", result)
	}
}

func TestParseN12(t *testing.T) {
	text := `текст
	<ul>
	  <li>текст</li>
	  <li>текст</li>
	  <b>текст</b>
	  <br>
	</ul>
	текст`

	result, _ := qvx.Parse(text)

	expect := `текст<br>` + "\n"
	expect += `<ul>` + "\n"
	expect += `<li>текст</li>` + "\n"
	expect += `<li>текст</li>` + "\n"
	expect += `</ul>` + "\n"
	expect += `текст`

	if result != expect {
		t.Errorf("Expect result to equal in func TestParseN12(t *testing.T).\n%s", result)
	}
}

func TestParseN13(t *testing.T) {
	text := `текст
	<li>текст</li>
	<li>текст</li>
	текст`

	result, _ := qvx.Parse(text)

	expect := `текст<br>` + "\n"
	expect += `текст<br>` + "\n"
	expect += `текст<br>` + "\n"
	expect += `текст`

	if result != expect {
		t.Errorf("Expect result to equal in func TestParseN13(t *testing.T).\n%s", result)
	}
}

func TestParseN14(t *testing.T) {
	text := `<b>"текст" текст "текст "текст" текст" "..."</b>`

	result, _ := qvx.Parse(text)

	expect := `<b>«текст» текст «текст „текст“ текст» «...»</b>`

	if result != expect {
		t.Errorf("Expect result to equal in func TestParseN14(t *testing.T).\n%s", result)
	}
}

func TestParseN15(t *testing.T) {
	text := `<pre>"текст" текст "текст "текст" текст" "..."</pre>`

	result, _ := qvx.Parse(text)

	expect := `<pre>&#34;текст&#34; текст &#34;текст &#34;текст&#34; текст&#34; &#34;...&#34;</pre>`

	if result != expect {
		t.Errorf("Expect result to equal in func TestParseN15(t *testing.T).\n%s", result)
	}
}

func TestParseN16(t *testing.T) {
	text := `текст &#40; &#41; &#42; &#43; &#44; текст`

	result, _ := qvx.Parse(text)

	expect := `текст ( ) * + , текст`

	if result != expect {
		t.Errorf("Expect result to equal in func TestParseN16(t *testing.T).\n%s", result)
	}
}

func TestParseN17(t *testing.T) {
	text := `текст - текст`

	result, _ := qvx.Parse(text)

	expect := `текст — текст`

	if result != expect {
		t.Errorf("Expect result to equal in func TestParseN17(t *testing.T).\n%s", result)
	}
}

func TestParseN18(t *testing.T) {
	text := `текст #hash и #tagname!`

	result, _ := qvx.Parse(text)

	expect := `текст <a href="/tags/hash/">#hash</a> и <a href="/tags/tagname/">#tagname</a>!`

	if result != expect {
		t.Errorf("Expect result to equal in func TestParseN18(t *testing.T).\n%s", result)
	}
}

func TestParseN19(t *testing.T) {
	text := `текст <b>#hash, #tagname</b> текст`

	result, _ := qvx.Parse(text)

	expect := `текст <b><a href="/tags/hash/">#hash</a>, <a href="/tags/tagname/">#tagname</a></b> текст`

	if result != expect {
		t.Errorf("Expect result to equal in func TestParseN19(t *testing.T).\n%s", result)
	}
}

func TestParseN20(t *testing.T) {
	text := `текст <a href="http://dighub.ru">DigHub</a> текст`

	result, _ := qvx.Parse(text)

	expect := `текст <a href="http://dighub.ru" rel="nofollow">DigHub</a> текст`

	if result != expect {
		t.Errorf("Expect result to equal in func TestParseN20(t *testing.T).\n%s", result)
	}
}

func TestParseN21(t *testing.T) {
	text := `текст <a href = "http://dighub.ru" title="text" >DigHub</a> текст`

	result, _ := qvx.Parse(text)

	expect := `текст <a href="http://dighub.ru" title="text" rel="nofollow">DigHub</a> текст`

	if result != expect {
		t.Errorf("Expect result to equal in func TestParseN21(t *testing.T).\n%s", result)
	}
}

func TestParseN22(t *testing.T) {
	text := `текст <a href="http://dighub.ru" args="test">DigHub</a> текст`

	result, _ := qvx.Parse(text)

	expect := `текст <a href="http://dighub.ru" rel="nofollow">DigHub</a> текст`

	if result != expect {
		t.Errorf("Expect result to equal in func TestParseN22(t *testing.T).\n%s", result)
	}
}

func TestParseN23(t *testing.T) {
	text := `текст <a href=http://dighub.ru title = text rel="nofollow">DigHub</a> текст`

	result, _ := qvx.Parse(text)

	expect := `текст <a href="http://dighub.ru" title="text" rel="nofollow">DigHub</a> текст`

	if result != expect {
		t.Errorf("Expect result to equal in func TestParseN23(t *testing.T).\n%s", result)
	}
}

func TestParseN24(t *testing.T) {
	text := `текст <a href=http://dighub.ru title=/" target=_blank>DigHub</a> текст`

	result, _ := qvx.Parse(text)

	expect := `текст <a href="http://dighub.ru" title="/&#34;" target="_blank" rel="nofollow">DigHub</a> текст`

	if result != expect {
		t.Errorf("Expect result to equal in func TestParseN24(t *testing.T).\n%s", result)
	}
}

func TestParseN25(t *testing.T) {
	text := `текст <a href="javascript:alert(1)">текст</a> текст`

	result, _ := qvx.Parse(text)

	expect := `текст текст текст`

	if result != expect {
		t.Errorf("Expect result to equal in func TestParseN25(t *testing.T).\n%s", result)
	}
}

func TestParseN26(t *testing.T) {
	text := `<b>текст текст</b> <cut> <b>текст <cut> текст</b>`

	result, _ := qvx.Parse(text)

	expect := `<b>текст текст</b> <cut> <b>текст текст</b>`

	if result != expect {
		t.Errorf("Expect result to equal in func TestParseN26(t *testing.T).\n%s", result)
	}
}
