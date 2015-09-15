
## Документации по конфигурации

### cfgAllowTags

cfgAllowTags — Задает список разрешенных тегов

`cfgAllowTags(tags []string)`

**Параметры**
* tags []string — теги

**Пример использования**
```go
qvx.cfgAllowTags([]string{"b", "i", "u", "a", "img", "ul", "ol", "li", "br", "code", "cut"})
```

### cfgSetTagShort

cfgSetTagShort — Указывает какие теги считать короткими

`qvx.cfgSetTagShort(tags []string)`

**Параметры**
* tags []string — теги

**Пример использования**
```go
qvx.cfgSetTagShort([]string{"br", "img", "cut"})
```

### cfgSetTagPreformatted

cfgSetTagPreformatted — Указывает преформатированные теги, в которых нужно всё заменять на HTML сущности

`qvx.cfgSetTagPreformatted(tags []string)`

**Параметры**
* tags []string — теги

**Пример использования**
```go
qvx.cfgSetTagPreformatted([]string{"code"})
```

### cfgSetTagNoTypography

cfgSetTagNoTypography — Указывает теги в которых нужно отключить типографирование текста

`qvx.cfgSetTagNoTypography(tags []string)`

**Параметры**
* tags []string — теги

**Пример использования**
```go
qvx.cfgSetTagNoTypography([]string{"code", "pre"})
```

### cfgSetTagIsEmpty

cfgSetTagIsEmpty — Указывает не короткие теги, которые могут быть пустыми и их не нужно из-за этого удалять

`qvx.cfgSetTagIsEmpty(tags []string)`

**Параметры**
* tags []string — теги

**Пример использования**
```go
qvx.cfgSetTagIsEmpty([]string{"div"})
```

### cfgSetTagNoAutoBr

cfgSetTagNoAutoBr — Указывает теги внутри, которых не нужна авто-расстановка тегов перевода на новую строку

`qvx.cfgSetTagNoAutoBr(tags []string)`

**Параметры**
* tags []string — теги

**Пример использования**
```go
qvx.cfgSetTagNoAutoBr([]string{"ul", "ol"})
```

### cfgSetTagCutWithContent

cfgSetTagCutWithContent — Указывает теги, которые необходимо вырезать вместе с содержимым

`qvx.cfgSetTagCutWithContent(tags []string)`

**Параметры**
* tags []string — теги

**Пример использования**
```go
qvx.cfgSetTagCutWithContent([]string{"script", "object", "iframe", "style"})
```

### cfgSetTagBlockType

cfgSetTagBlockType — Указывает теги после, которых не нужно добавлять дополнительный перевод строки, например, блочные теги

`qvx.cfgSetTagBlockType(tags []string)`

**Параметры**
* tags []string — теги

**Пример использования**
```go
qvx.cfgSetTagBlockType([]string{"ol", "ul", "code"})
```

### cfgAllowTagParams

cfgAllowTagParams — Добавляет разрешенные параметры для тегов.

`qvx.cfgAllowTagParams(tag string, params []string)`

**Параметры**
* tag string — тег
* params []string — разрешённые параметры

**Пример использования**
```go
qvx.CfgAllowTagParams("a", []string{"href", "title", "target", "rel"})
qvx.CfgAllowTagParams("img", []string{"src", "alt", "title", "align", "width", "height"})
```

### CfgSetTagParamsRequired

CfgSetTagParamsRequired — Добавляет обязательные параметры для тега, без которых тег будет удален

`qvx.cfgSetTagParamsRequired(tag string, params []string)`

**Параметры**
* tag string — тег
* params []string — обязательные параметры

**Пример использования**
```go
qvx.CfgSetTagParamsRequired("a", []string{"href"})
qvx.CfgSetTagParamsRequired("img", []string{"src"})
```

### CfgAllowTagParamValue

CfgAllowTagParamValue — Уточняет значения параметра тега.
Значение по умолчанию - шаблон #str. Разрешенные шаблоны #str, #int, #link, #regexp(...).
Например, шаблон с регулярным выражением может выглядеть так: "#regexp(\d+(%|px))"

`qvx.CfgAllowTagParamValue(tag string, param string, value interface{})`

**Параметры**
* tag string — тег
* param string — параметр
* value interface{} — значение параметра, может быть строка или срез строк, разрешены шаблоны #str, #int, #link, #regexp(...)

**Пример использования**
```go
qvx.CfgAllowTagParamValue("a", "href", "#link")
qvx.CfgAllowTagParamValue("a", "target", "_blank")

qvx.CfgAllowTagParamValue("img", "align", []string{"right", "left", "center"})
qvx.CfgAllowTagParamValue("img", "width", "#int")
qvx.CfgAllowTagParamValue("img", "height", "#int")
```

### CfgSetTagParamDefault

CfgSetTagParamDefault — Указывает значения по умолчанию для параметров тега.
Если параметры у тега отсутствуют, то они будут добавлены автоматически.

`qvx.CfgSetTagParamDefault(tag string, param string, value string)`

**Параметры**
* tag string — тег
* param string — параметр
* value string — значение

**Пример использования**
```go
qvx.CfgSetTagParamDefault("a", "rel", "nofollow")
qvx.CfgSetTagParamDefault("img", "alt", "")
```

### CfgSetTagParamReview

CfgSetTagParamReview — Указывает параметры значение которых нужно заменять на указанные значения
Заменит значение параметра на указанное, если оно отличается.

`qvx.CfgSetTagParamReview(tag string, param string, value string)`

**Параметры**
* tag string — тег
* param string — параметр
* value string — значение

**Пример использования**
```go
qvx.CfgSetTagParamReview("a", "rel", "nofollow")
```

### cfgSetTagChilds

cfgSetTagChilds — Указывает какие теги являются контейнерами для других тегов

`qvx.cfgSetTagChilds(tag string, childs []string)`

**Параметры**
* tag string — тег
* childs []string — разрешённые дочерние теги

**Пример использования**
```go
qvx.CfgSetTagChilds("ul", []string{"li"})
qvx.CfgSetTagChilds("ol", []string{"li"})
```

### CfgSetTagParentOnly

CfgSetTagParentOnly — Указывает, какие теги могут быть только контейнерами для других тегов

`qvx.CfgSetTagParentOnly(tags []string)`

**Параметры**
* tags []string — теги являются только контейнером для других тегов и не могут содержать текст

**Пример использования**
```go
qvx.CfgSetTagParentOnly([]string{"ul", "ol"})
```

### CfgSetTagChildOnly

CfgSetTagChildOnly — Указывает, какие теги могут быть только дочерними для других тегов

`qvx.CfgSetTagChildOnly(tags []string)`

**Параметры**
* tags []string — теги являются только дочерними для других тегов

**Пример использования**
```go
qvx.CfgSetTagChildOnly([]string{"li"})
```

### cfgSetTagGlobal

cfgSetTagGlobal — Указывает какие теги не должны быть дочерними к другим тегам

`qvx.cfgSetTagGlobal(tags []string)`

**Параметры**
* tags []string — теги

**Пример использования**
```go
qvx.cfgSetTagGlobal([]string{"cut"})
```

### cfgSetTagBuildCallback

cfgSetTagBuildCallback — Устанавливает на тег callback-функцию для построения тега

`qvx.cfgSetTagBuildCallback(tag string, callback func(string, map[string]string, string) string)`

**Параметры**
* tag string — тег
* callback func(string, map[string]string, string) string — функция

**Пример использования**
```go
qvx.cfgSetTagBuildCallback("code", TagCodeBuild)

//...

func TagCodeBuild(tag string, params map[string]string, content string) string {
	return "<pre><code>" + content + "<code><pre>\n"
}
```

### cfgSetSpecialCharCallback

cfgSetSpecialCharCallback — Устанавливает на строку предваренную спецсимволом callback-функцию. По умолчанию Qevix работает с тремя спец. символами #, @, $.
Как можно догадаться, эта настройка позволяет получить хештег (#tagname), имя пользователя (@username), или ключевое слово ($keyword) и оформить его в виде ссылки или того, что вам нужно.

`cfgSetSpecialCharCallback(char rune, callback func(string) string)`

**Параметры**
* char rune — спецсимвол #, @, $
* callback func(string)string — функция

**Пример использования**
```go
qvx.CfgSetSpecialCharCallback('#', TagSharpBuild)
qvx.CfgSetSpecialCharCallback('@', TagAtBuild)

//...

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
```

Вы можете сами отслеживать, какие символы могут входить в строку.

### cfgSetLinkProtocolAllow

cfgSetLinkProtocolAllow — Устанавливает список разрешенных протоколов для ссылок. По умолчанию разрешены http, https, ftp

`cfgSetLinkProtocolAllow(protocols []string)`

**Параметры**
* protocols []string — срез протоколов

**Пример использования**
```go
qvx.cfgSetLinkProtocolAllow([]string{"http", "https"})
```

### cfgSetXHTMLMode

cfgSetXHTMLMode — Включает или выключает режим XHTML. По умолчанию выключен.

`cfgSetXHTMLMode(isXHTMLMode bool)`

**Параметры**
* isXHTMLMode bool — Включить XHTML формат тегов установив в True;

**Пример использования**
```go
qvx.cfgSetXHTMLMode(true)
```

### cfgSetAutoBrMode

cfgSetAutoBrMode — Включает или выключает режим автозамены символов перевода строки на тег br. По умолчанию включен.

`cfgSetAutoBrMode(isAutoBrMode bool)`

**Параметры**
* isAutoBrMode bool — Включить авторасстановку тегов переда строки установив в True;

**Пример использования**
```go
qvx.cfgSetAutoBrMode(true)
```

### cfgSetAutoLinkMode

cfgSetAutoLinkMode — Включает или выключает режим автоматического определения ссылок. По умолчанию режим включен.

`cfgSetAutoLinkMode(isAutoLinkMode bool)`

**Параметры**
* isAutoLinkMode bool — Включить автоопределение ссылок установив в True;

**Пример использования**
```go
qvx.cfgSetAutoLinkMode(true)
```

### cfgSetEOL

cfgSetEOL — Задает символ/символы перевода строки для текста на выходе. По умолчанию используется только символ перевода строки (LF) "\n", можно задать (CR+LF) "\r\n".

`cfgSetEOL(nl string)`

**Параметры**
* nl string — "\n" или "\r\n"

**Пример использования**
```go
qvx.cfgSetEOL("\r\n")
```
