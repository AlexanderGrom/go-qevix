
## Qevix

**Qevix** — Aвтоматический фильтр HTML/XHTML разметки в текстах.
Применяя наборы правил, контролирует перечень допустимых тегов и атрибутов, предотвращает возможные XSS-атаки.
Oсновывается на [PHP версии Qevix](https://github.com/AlexanderGrom/php-qevix/), но не повторяет её полностью.

### Возможности

* Фильтрация текста с HTML/XHTML разметкой на основе заданных правил о разрешённых тегах и атрибутах;
* Исправление ошибок HTML/XHTML;
* Обработка строк предваренных специальными символами (#tagname, @username, $keyword);
* Установка на теги callback-функций;
* Предотвращение XSS-атак;

### Пример использования

```go
package main

import (
	"ftm"
	"net/url"
	"qevix"
	"regexp"
)

func main() {
	// Инициализация
	qvx := qevix.New()

	// Конфигурация

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
	qvx.CfgSetTagBlockType([]string{"ol", "ul", "code", "div"})

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

	//Фильтр

	text := `<b>Жирный</b>... <b><i>Жирный курсив</i></b>
	<!-- Удалить комментарий -->
	<h1>Тег не разрешен</h1>

	<!-- Ниже пустой тег, который будет удален -->
	<p></p>
	<!-- Тег ниже будет оставлен -->
	<div></div>

	Метки или теги #qevix, #golang, #parser.
	Люди @Alexander, @Андрей, @Семен_Семеныч.

	<s>Зачеркивать нельзя, <u>a подчеркивать можно</u></s>,
	<script type="text/javascript">alert('Это будет удалено')</script>
	Теги iframe, style и object будут удалены согласно тикущим правилам.

	<b>Одинаковые <b>вложенные</b> теги</b>

	Тег ниже обрабатывается callback функцией:

	<code>
		<body>
			<b>JavaScript:</b>
			<script>alert('Hello World')</script>
		</body>
	</code>

	Подсветка ссылок http://dighub.ru или www.yandex.ru!
	Подсветка в скобках (https://github.com)!
	Подсветка в теге <b>http://webonrails.ru</b>.
	Более сложно  http://yandex.ru/yandsearch?lr=2&text=qevix!..
	Тегом <a href="https://ru.wikipedia.org">Википедия</a>
	Без JavaScript <a href="javascript:alert('Hi!')" title="Нажми на меня">Hello World!</a>
	Без ненужных атрибутов <a href="https://github.com" name="top" hreflang="ru">GitHub</a>
	С кривыми атрибутами <a href=http://golang.org title = text target=_blank>Атрибуты без кавычек!</a>

	Изображения <img src="http://php.net/images/news/phpday2012.png" alt="Image">

	"кавычки" елочки и "соблюдение "вложенности" кавычек"...

	Использование преформатирования <pre>"кавычки" елочки и "соблюдение "вложенности" кавычек"</pre>

	Теги могут быть только внутри родительских
	<li>Пункт 1</li>
	<li>Пункт 2</li>

	Правильно:

	<ul>
	  <li>Пункт 1</li>
	  <li>Пункт 2</li>
	</ul>

	Преобразовать коды символов &#40; &#41; &#42; &#43; &#44; обратно в символы.

	Преобразовать в тексте короткое тире - в длинное, но не в таком (2-2=0) и не в таком (веб-программирование)

	Работа с пунктами (диалогами):
	- Пункт 1
	- Пункт 2
	- Пункт 3

	Правильный CUT: <b>Краткая часть</b> <cut> <b>Тег <cut> не может быть вложенным</b>

	Нужно    убирать    лишние пробелы и самому <b>закрыть <u>теги`

	result, _ := qvx.Parse(text)

	fmt.Print(result)
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
```

### Документация по конфигурации

* [DOCUMENTATION](DOCUMENTATION.md)

### Поддержка

* **Александр Громов** - пишите в [Issues](https://github.com/AlexanderGrom/go-qevix/issues)

------

Дайте мне знать, если вы нашли проблему в **Qevix** или вас не устраивает его работа.
Спасибо!
