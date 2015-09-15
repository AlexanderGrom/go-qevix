// Qevix - Автоматический фильтр HTML/XHTML разметки в тексте.
// Основан на PHP версии Qevix.
//
// Возможности:
// - Фильтрация текста с HTML/XHTML разметкой на основе заданных правил о разрешённых тегах и атрибутах;
// - Исправление ошибок HTML/XHTML;
// - Обработка строк предваренных специальными символами (#tagname, @username, $keyword);
// - Установка на теги callback-функций;
// - Предотвращение XSS-атак;
//

package qevix

import (
	"bytes"
	"errors"
	"html"
	"regexp"
	"strings"
)

const (
	NULL           = 0x0
	PRINATABLE     = 0x1
	ALPHA          = 0x2
	NUMERIC        = 0x4
	PUNCTUATUON    = 0x8
	SPACE          = 0x10
	NL             = 0x20
	TAG_NAME       = 0x40
	TAG_PARAM_NAME = 0x80
	TAG_QUOTE      = 0x100
	TEXT_QUOTE     = 0x200
	TEXT_BRACKET   = 0x400
	SPECIAL_CHAR   = 0x800
	//const = 0x1000
	//const = 0x2000
	//const = 0x4000
	//const = 0x8000
	NOPRINT = 0x10000
)

// Сгенерированные классы символов (не трогать!)
var CHAR_CLASSES = map[rune]int{
	0: 65536, 1: 65536, 2: 65536, 3: 65536, 4: 65536, 5: 65536, 6: 65536, 7: 65536, 8: 65536,
	9: 65552, 10: 65568, 11: 65536, 12: 65536, 13: 65568, 14: 65536, 15: 65536, 16: 65536,
	17: 65536, 18: 65536, 19: 65536, 20: 65536, 21: 65536, 22: 65536, 23: 65536, 24: 65536,
	25: 65536, 26: 65536, 27: 65536, 28: 65536, 29: 65536, 30: 65536, 31: 65536, 32: 65552,
	97: 195, 98: 195, 99: 195, 100: 195, 101: 195, 102: 195, 103: 195, 104: 195, 105: 195,
	106: 195, 107: 195, 108: 195, 109: 195, 110: 195, 111: 195, 112: 195, 113: 195, 114: 195,
	115: 195, 116: 195, 117: 195, 118: 195, 119: 195, 120: 195, 121: 195, 122: 195, 65: 195,
	66: 195, 67: 195, 68: 195, 69: 195, 70: 195, 71: 195, 72: 195, 73: 195, 74: 195, 75: 195,
	76: 195, 77: 195, 78: 195, 79: 195, 80: 195, 81: 195, 82: 195, 83: 195, 84: 195, 85: 195,
	86: 195, 87: 195, 88: 195, 89: 195, 90: 195, 48: 197, 49: 197, 50: 197, 51: 197, 52: 197,
	53: 197, 54: 197, 55: 197, 56: 197, 57: 197, 45: 129, 34: 769, 39: 257, 46: 9, 44: 9, 33: 9,
	63: 9, 58: 9, 59: 9, 60: 1025, 62: 1025, 91: 1025, 93: 1025, 123: 1025, 125: 1025, 40: 1025,
	41: 1025, 64: 2049, 35: 2049, 36: 2049,
}

//
// Состояние (точка востановления) на головке автомата
//
type state struct {
	prevPos       int  // Позиция в последовательности символов
	prevChar      rune // Символ (руна)
	prevCharClass int  // Класс символа

	curPos       int
	curChar      rune
	curCharClass int

	nextPos       int
	nextChar      rune
	nextCharClass int
}

//
// Парсер
// qvx := qevix.New()
// qvx.cnfTagAllow([]string{"a","b","i"})
// str, err := qvx.Parse(text)
//
type parser struct {
	tagAllowed map[string]bool // Тег допустим

	tagParamAllowed  map[string]map[string][]string // Параметр тега допустим
	tagParamRequired map[string]map[string]bool     // Параметр тега является необходимым

	tagParamSorted map[string][]string // Отсортированый срез имен параметров для тегов

	tagParamDefault map[string]map[string]string // Автодобавление параметров со значениями по умолчанию
	tagParamReview  map[string]map[string]string // Параметры значения которых нужно заменить на указанные

	tagShort          map[string]bool // Тег короткий
	tagCutWithContent map[string]bool // Тег необходимо вырезать вместе с его контентом
	tagGlobalOnly     map[string]bool // Тег может находиться только в "глобальной" области видимости (не быть дочерним к другим)
	tagParentOnly     map[string]bool // Тег может содержать только другие теги
	tagChildOnly      map[string]bool // Тег может находиться только внутри других тегов

	tagParent map[string]map[string]bool // Тег родитель относительно дочернего тега
	tagChild  map[string]map[string]bool // Тег дочерний относительно родительского

	tagPreformatted map[string]bool // Преформатированные теги
	tagNoTypography map[string]bool // Тег с отключенным типографированием
	tagEmpty        map[string]bool // Пустой не короткий тег
	tagNoAutoBr     map[string]bool // Тег в котором не нужна авто-расстановка <br>
	tagBlockType    map[string]bool // Тег после которого нужно удалять один перевод строки

	tagBuildCallback map[string]func(string, map[string]string, string) string // Тег обрабатывается и строится callback-функцией

	entities    map[rune]string // Сепецсимволы для замены на HTML эквиваленты
	quotes      [][]rune        // Замена кавычек
	bracketsALL map[rune]rune   // Скобки
	dash        string          // Тире
	nl          string          // Символы перевода строки
	br          string          // Тег <br>

	textBuf []rune // Буфер с рунами
	textLen int    // Длина буфера рун

	prevPos       int  // Предыдущая позиция символа
	prevChar      rune // Предыдущий символ
	prevCharClass int  // Предыдущий класс символа

	curPos       int  // Текущая позиция символа
	curChar      rune // Текущий символ
	curCharClass int  // Текущий класс символа

	nextPos       int  // Следующая позиция символа
	nextChar      rune // Следующий символ
	nextCharClass int  // Следующий класс символа

	curTag            string   // Текущий тег
	statesStack       []state  // Стек состояний
	quotesOpened      int      // Кол-во открытых кавычек
	linkProtocolAllow []string // Разрешенные схемы для ссылок

	specialChars map[rune]func(string) string // Функции повешенные на специальные символы (@,#,$)

	isXHTMLMode       bool // Включение режима XHTML
	isAutoBrMode      bool // Включение авторасстановки тегов переноса строк
	isAutoLinkMode    bool // Включение автоподсветки ссылок
	isSpecialCharMode bool // Включение отлавливания строк предваренных специальными символами (@,#,$)
	isTypoMode        bool // Влючение типографирования

	errorsList []error // Ошибки в разметке произошедшие за время парсинга
}

func New() *parser {
	return &parser{
		tagAllowed: make(map[string]bool),

		tagParamAllowed:  make(map[string]map[string][]string),
		tagParamRequired: make(map[string]map[string]bool),

		tagParamSorted: make(map[string][]string),

		tagParamDefault: make(map[string]map[string]string),
		tagParamReview:  make(map[string]map[string]string),

		tagShort:          make(map[string]bool),
		tagCutWithContent: make(map[string]bool),
		tagGlobalOnly:     make(map[string]bool),
		tagParentOnly:     make(map[string]bool),
		tagChildOnly:      make(map[string]bool),

		tagParent: make(map[string]map[string]bool),
		tagChild:  make(map[string]map[string]bool),

		tagPreformatted: make(map[string]bool),
		tagNoTypography: make(map[string]bool),
		tagEmpty:        make(map[string]bool),
		tagNoAutoBr:     make(map[string]bool),
		tagBlockType:    make(map[string]bool),

		tagBuildCallback: make(map[string]func(string, map[string]string, string) string),

		entities: map[rune]string{
			'"': "&#34;", '\'': "&#39;", '<': "&#60;", '>': "&#62;", '&': "&#38;",
		},
		quotes: [][]rune{
			[]rune{'«', '»'}, []rune{'„', '“'},
		},
		bracketsALL: map[rune]rune{
			'<': '>', '[': ']', '{': '}', '(': ')',
		},

		dash: "—",
		nl:   "\n",
		br:   "<br>",

		textBuf: []rune{},
		textLen: 0,

		prevPos:       -1,
		prevChar:      -1,
		prevCharClass: NULL,

		curPos:       -1,
		curChar:      -1,
		curCharClass: NULL,

		nextPos:       -1,
		nextChar:      -1,
		nextCharClass: NULL,

		curTag:       "",
		statesStack:  []state{},
		quotesOpened: 0,
		linkProtocolAllow: []string{
			"http", "https", "ftp",
		},
		specialChars: make(map[rune]func(string) string),

		isXHTMLMode:       false,
		isAutoBrMode:      true,
		isAutoLinkMode:    true,
		isSpecialCharMode: false,
		isTypoMode:        true,

		errorsList: []error{},
	}
}

//
// Парсинг строки
//
// text string - входная строка для парсинга
//
func (self *parser) Parse(text string) (string, []error) {
	// Обнуляем параметры
	self.prevPos = -1
	self.prevChar = 0
	self.prevCharClass = NULL

	self.curPos = -1
	self.curChar = 0
	self.curCharClass = NULL

	self.nextPos = -1
	self.nextChar = 0
	self.nextCharClass = NULL

	self.curTag = ""

	self.statesStack = []state{}

	self.quotesOpened = 0

	text = strings.Replace(text, "\r", "", -1)

	self.textBuf = []rune(text)
	self.textLen = len(self.textBuf)

	self.errorsList = []error{}

	self.movePos(0)

	content := ""
	content = self.makeContent("")
	content = strings.Replace(content, "\n", self.nl, -1)
	content = strings.TrimSpace(content)

	errors := self.errorsList

	return content, errors
}

//
// КОНФИГУРАЦИЯ: Задает список разрешенных тегов
//
// tags []string - теги
//
func (self *parser) CfgAllowTags(tags []string) {
	for _, tag := range tags {
		self.tagAllowed[tag] = true
	}
}

//
// КОНФИГУРАЦИЯ: Указывает, какие теги считать короткими (<br>, <img>)
//
// tags []string - теги
//
func (self *parser) CfgSetTagShort(tags []string) {
	for _, tag := range tags {
		if _, ok := self.tagAllowed[tag]; !ok {
			panic("Тег '" + tag + "' отсутствует в списке разрешённых тегов")
		}
		self.tagShort[tag] = true
	}
}

//
// КОНФИГУРАЦИЯ: Указывает преформатированные теги, в которых нужно всё заменять на HTML сущности
//
// tags []string - теги
//
func (self *parser) CfgSetTagPreformatted(tags []string) {
	for _, tag := range tags {
		if _, ok := self.tagAllowed[tag]; !ok {
			panic("Тег '" + tag + "' отсутствует в списке разрешённых тегов")
		}
		self.tagPreformatted[tag] = true
	}
}

//
// КОНФИГУРАЦИЯ: Указывает теги в которых нужно отключить типографирование текста
//
// tags []string - теги
//
func (self *parser) CfgSetTagNoTypography(tags []string) {
	for _, tag := range tags {
		if _, ok := self.tagAllowed[tag]; !ok {
			panic("Тег '" + tag + "' отсутствует в списке разрешённых тегов")
		}
		self.tagNoTypography[tag] = true
	}
}

//
// КОНФИГУРАЦИЯ: Указывает не короткие теги, которые могут быть пустыми и их не нужно из-за этого удалять
//
// tags []string - теги
//
func (self *parser) CfgSetTagIsEmpty(tags []string) {
	for _, tag := range tags {
		if _, ok := self.tagAllowed[tag]; !ok {
			panic("Тег '" + tag + "' отсутствует в списке разрешённых тегов")
		}
		self.tagEmpty[tag] = true
	}
}

//
// КОНФИГУРАЦИЯ: Указывает теги внутри которых не нужна авторасстановка тегов перевода на новую строку
//
// tags []string - теги
//
func (self *parser) CfgSetTagNoAutoBr(tags []string) {
	for _, tag := range tags {
		if _, ok := self.tagAllowed[tag]; !ok {
			panic("Тег '" + tag + "' отсутствует в списке разрешённых тегов")
		}
		self.tagNoAutoBr[tag] = true
	}
}

//
// КОНФИГУРАЦИЯ: Указывает теги, которые необходимо вырезать вместе с содержимым (style, script, iframe)
//
// tags []string - теги
//
func (self *parser) CfgSetTagCutWithContent(tags []string) {
	for _, tag := range tags {
		self.tagCutWithContent[tag] = true
	}
}

//
// КОНФИГУРАЦИЯ: Указывает теги после которых не нужно добавлять дополнительный перевод строки, например, блочные теги
//
// tags []string - теги
//
func (self *parser) CfgSetTagBlockType(tags []string) {
	for _, tag := range tags {
		if _, ok := self.tagAllowed[tag]; !ok {
			panic("Тег '" + tag + "' отсутствует в списке разрешённых тегов")
		}
		self.tagBlockType[tag] = true
	}
}

//
// КОНФИГУРАЦИЯ: Добавляет разрешенные параметры для тегов
//
// tag string - тег
// params []string - разрешённые параметры
//
func (self *parser) CfgAllowTagParams(tag string, params []string) {
	if _, ok := self.tagAllowed[tag]; !ok {
		panic("Тег '" + tag + "' отсутствует в списке разрешённых тегов")
	}
	self.tagParamAllowed[tag] = make(map[string][]string)
	self.tagParamSorted[tag] = []string{}
	for _, param := range params {
		self.tagParamAllowed[tag][param] = []string{"#str"}
		self.tagParamSorted[tag] = append(self.tagParamSorted[tag], param)
	}
}

//
// КОНФИГУРАЦИЯ: Добавляет обязательные параметры для тега
//
// tag string - тег
// params []string - обязательные параметры
//
func (self *parser) CfgSetTagParamsRequired(tag string, params []string) {
	if _, ok := self.tagAllowed[tag]; !ok {
		panic("Тег '" + tag + "' отсутствует в списке разрешённых тегов")
	}
	self.tagParamRequired[tag] = make(map[string]bool)
	for _, param := range params {
		self.tagParamRequired[tag][param] = true
	}
}

//
// КОНФИГУРАЦИЯ: Уточняет значения параметра тега
//
// tag string - тег
// param string - параметр
// value interface{} - значение параметра, может быть строка или срез строк, разрешены шаблоны #str, #int, #link, #regexp(...)
//
func (self *parser) CfgAllowTagParamValue(tag string, param string, value interface{}) {
	if _, ok := self.tagAllowed[tag]; !ok {
		panic("Тег '" + tag + "' отсутствует в списке разрешённых тегов")
	}
	if _, ok := self.tagParamAllowed[tag][param]; !ok {
		panic("Параметр '" + param + "' тега '" + tag + "' отсутствует в списке разрешённых параметров")
	}
	var val []string
	switch v := value.(type) {
	case string:
		val = []string{v}
	case []string:
		val = v
	default:
		panic("CfgAllowTagParamValue: тег " + tag + ", параметр " + param + ", значение должно быть строкой или срезом строк")
	}
	self.tagParamAllowed[tag][param] = val
}

//
// КОНФИГУРАЦИЯ: Указывает, какие теги являются контейнерами для других тегов
//
// tag string - тег
// childs []string - разрешённые дочерние теги
//
func (self *parser) CfgSetTagChilds(tag string, childs []string) {
	if _, ok := self.tagAllowed[tag]; !ok {
		panic("Тег '" + tag + "' отсутствует в списке разрешённых тегов")
	}
	self.tagChild[tag] = make(map[string]bool)
	for _, child := range childs {
		if _, ok := self.tagAllowed[child]; !ok {
			panic("Тег '" + tag + "' отсутствует в списке разрешённых тегов")
		}
		if _, ok := self.tagParent[child]; !ok {
			self.tagParent[child] = make(map[string]bool)
		}
		self.tagChild[tag][child] = true
		self.tagParent[child][tag] = true
	}
}

//
// КОНФИГУРАЦИЯ: Указывает, какие теги могут быть только контейнерами для других тегов
//
// tags []string - теги являются только контейнером для других тегов и не могут содержать текст
//
func (self *parser) CfgSetTagParentOnly(tags []string) {
	for _, tag := range tags {
		if _, ok := self.tagAllowed[tag]; !ok {
			panic("Тег '" + tag + "' отсутствует в списке разрешённых тегов")
		}
		self.tagParentOnly[tag] = true
	}
}

//
// КОНФИГУРАЦИЯ: Указывает, какие теги могут быть только дочерними для других тегов
//
// tags []string - теги являются только дочерними для других тегов
//
func (self *parser) CfgSetTagChildOnly(tags []string) {
	for _, tag := range tags {
		if _, ok := self.tagAllowed[tag]; !ok {
			panic("Тег '" + tag + "' отсутствует в списке разрешённых тегов")
		}
		self.tagChildOnly[tag] = true
	}
}

//
// КОНФИГУРАЦИЯ: Указывает, какие теги не должны быть дочерними к другим тегам
//
// tags []string - теги
//
func (self *parser) CfgSetTagGlobal(tags []string) {
	for _, tag := range tags {
		if _, ok := self.tagAllowed[tag]; !ok {
			panic("Тег '" + tag + "' отсутствует в списке разрешённых тегов")
		}
		self.tagGlobalOnly[tag] = true
	}
}

//
// КОНФИГУРАЦИЯ: Указывает значения по умолчанию для параметров тега если они не заданы
//
// tag string - тег
// param string - параметр
// value string - значение
//
func (self *parser) CfgSetTagParamDefault(tag string, param string, value string) {
	if _, ok := self.tagAllowed[tag]; !ok {
		panic("Тег '" + tag + "' отсутствует в списке разрешённых тегов")
	}
	if _, ok := self.tagParamDefault[tag]; !ok {
		self.tagParamDefault[tag] = make(map[string]string)
	}
	self.tagParamDefault[tag][param] = value
}

//
// КОНФИГУРАЦИЯ: Указывает параметры значение которых нужно заменять на указанные значения
//
// tag string - тег
// param string - параметр
// value string - значение
//
func (self *parser) CfgSetTagParamReview(tag string, param string, value string) {
	if _, ok := self.tagAllowed[tag]; !ok {
		panic("Тег '" + tag + "' отсутствует в списке разрешённых тегов")
	}
	if _, ok := self.tagParamReview[tag]; !ok {
		self.tagParamReview[tag] = make(map[string]string)
	}
	self.tagParamReview[tag][param] = value
}

//
// КОНФИГУРАЦИЯ: Устанавливает на тег callback-функцию для построения тега
//
// tag string - тег
// callback func(string, map[string]string, string) string - функция
//
func (self *parser) CfgSetTagBuildCallback(tag string, callback func(string, map[string]string, string) string) {
	if _, ok := self.tagAllowed[tag]; !ok {
		panic("Тег '" + tag + "' отсутствует в списке разрешённых тегов")
	}
	self.tagBuildCallback[tag] = callback
}

//
// КОНФИГУРАЦИЯ: Устанавливает на строку предварённую спецсимволом callback-функцию
//
// char rune - спецсимвол
// callback func(string)string - функция
//
func (self *parser) CfgSetSpecialCharCallback(char rune, callback func(string) string) {
	if (self.getClassByOrd(char) & SPECIAL_CHAR) == NULL {
		panic("Значение параметр char метода CfgSetSpecialCharCallback отсутствует в списке разрешенных символов")
	}
	self.isSpecialCharMode = true
	self.specialChars[char] = callback
}

//
// КОНФИГУРАЦИЯ: Устанавливает список разрешенных протоколов для ссылок (https, http, ftp)
//
// protocols []string - список протоколов
//
func (self *parser) CfgSetLinkProtocolAllow(protocols []string) {
	self.linkProtocolAllow = protocols
}

//
// КОНФИГУРАЦИЯ: Включает или выключает режим XHTML
//
func (self *parser) CfgSetXHTMLMode(isXHTMLMode bool) {
	if isXHTMLMode {
		self.br = "<br/>"
	} else {
		self.br = "<br>"
	}
	self.isXHTMLMode = isXHTMLMode
}

//
// КОНФИГУРАЦИЯ: Включает или выключает режим автозамены символов переводов строк на тег <br>
//
func (self *parser) CfgSetAutoBrMode(isAutoBrMode bool) {
	self.isAutoBrMode = isAutoBrMode
}

//
// КОНФИГУРАЦИЯ: Включает или выключает режим автоматического определения ссылок
//
func (self *parser) CfgSetAutoLinkMode(isAutoLinkMode bool) {
	self.isAutoLinkMode = isAutoLinkMode
}

//
// КОНФИГУРАЦИЯ: Задает символ/символы перевода строки "\n" или "\r\n"
//
// nl string - "\n" или "\r\n"
//
func (self *parser) CfgSetEOL(nl string) {
	if nl == "\n" || nl == "\r\n" {
		self.nl = nl
	}
}

//
// Возвращает класс символа по его коду
//
// ord rune - код символа
//
func (self *parser) getClassByOrd(ord rune) int {
	if _, ok := CHAR_CLASSES[ord]; ok {
		return CHAR_CLASSES[ord]
	}
	return PRINATABLE
}

//
// Получение следующего символа из входной строки
//
func (self *parser) moveNextPos() bool {
	return self.movePos(self.curPos + 1)
}

//
// Получение пердыдущего символа из входной строки
//
func (self *parser) movePrevPos() bool {
	return self.movePos(self.curPos - 1)
}

//
// Перемещает указатель на указанную позицию во входной строке и считывание символа
//
// position int - позиция в тексте
//
func (self *parser) movePos(position int) bool {
	prevPos := position - 1
	curPos := position
	nextPos := position + 1

	self.prevPos = prevPos

	if prevPos < self.textLen && prevPos >= 0 {
		self.prevChar = self.textBuf[prevPos]
		self.prevCharClass = self.getClassByOrd(self.prevChar)
	} else {
		self.prevChar = 0
		self.prevCharClass = NULL
	}

	self.curPos = curPos

	if curPos < self.textLen && curPos >= 0 {
		self.curChar = self.textBuf[curPos]
		self.curCharClass = self.getClassByOrd(self.curChar)
	} else {
		self.curChar = 0
		self.curCharClass = NULL
	}

	self.nextPos = nextPos

	if nextPos < self.textLen && nextPos >= 0 {
		self.nextChar = self.textBuf[nextPos]
		self.nextCharClass = self.getClassByOrd(self.nextChar)
	} else {
		self.nextChar = 0
		self.nextCharClass = NULL
	}

	if self.curChar == 0 {
		return false
	}

	return true
}

//
// Сохраняет текущее состояние автомата
//
func (self *parser) saveState() {
	ss := state{
		prevPos:       self.prevPos,
		prevChar:      self.prevChar,
		prevCharClass: self.prevCharClass,

		curPos:       self.curPos,
		curChar:      self.curChar,
		curCharClass: self.curCharClass,

		nextPos:       self.nextPos,
		nextChar:      self.nextChar,
		nextCharClass: self.nextCharClass,
	}

	self.statesStack = append(self.statesStack, ss)
}

//
// Восстанавливает последнее сохраненное состояние автомата
//
func (self *parser) restoreState() {
	if len(self.statesStack) == 0 {
		return
	}

	ss := self.statesStack[len(self.statesStack)-1]
	self.statesStack = self.statesStack[:len(self.statesStack)-1]

	self.prevPos = ss.prevPos
	self.prevChar = ss.prevChar
	self.prevCharClass = ss.prevCharClass

	self.curPos = ss.curPos
	self.curChar = ss.curChar
	self.curCharClass = ss.curCharClass

	self.nextPos = ss.nextPos
	self.nextChar = ss.nextChar
	self.nextCharClass = ss.nextCharClass
}

//
// Удаляет последнее сохраненное состояние
//
func (self *parser) removeState() {
	if len(self.statesStack) > 1 {
		self.statesStack = self.statesStack[:len(self.statesStack)-1]
	}
}

//
// Проверяет точное вхождение символа в текущей позиции
//
// char rune - символ (руна)
//
func (self *parser) matchChar(char rune) bool {
	return (self.curChar == char)
}

//
// Проверяет вхождение символа указанного класса в текущей позиции
//
// charClass int - класс символа
//
func (self *parser) matchCharClass(charClass int) bool {
	return (self.curCharClass & charClass) != NULL
}

//
// Проверяет точное совпадение строки в текущей позиции
//
// str string - строка
//
func (self *parser) matchStr(str string) bool {
	self.saveState()

	runes := []rune(str)
	lenght := len(runes)
	buff := make([]rune, lenght)

	for i := 0; i < lenght && self.curCharClass != NULL; i++ {
		buff[i] = self.curChar
		self.moveNextPos()
	}

	self.restoreState()

	return EqualSliceRune(buff, runes)
}

//
// Пропускает текст до нахождения указанного символа
//
// char rune - символ (руна) для поиска
//
func (self *parser) skipTextToChar(char rune) bool {
	for self.curChar != char && self.curCharClass != NULL {
		self.moveNextPos()
	}
	return (self.curCharClass != NULL)
}

//
// Пропускает текст до нахождения указанной строки
//
// str string - строка или символ для поиска
//
func (self *parser) skipTextToStr(str string) bool {
	runes := []rune(str)

	for self.curCharClass != NULL {
		if self.curChar == runes[0] {
			self.saveState()

			state := true
			for _, ord := range runes {
				if self.curCharClass == NULL {
					self.removeState()
					return false
				}

				if self.curChar != ord {
					state = false
					break
				}

				self.moveNextPos()
			}

			self.restoreState()

			if state {
				return true
			}
		}

		self.moveNextPos()
	}

	return false
}

//
// Пропускает строку если она начинается с текущей позиции
//
// self.skipTextToStr("-->") && self.skipStr("-->")
//
// str string - строка для пропуска
//
func (self *parser) skipStr(str string) bool {
	self.saveState()

	runes := []rune(str)
	state := true
	for _, ord := range runes {
		if self.curChar != ord || self.curCharClass == NULL {
			state = false
			break
		}

		self.moveNextPos()
	}

	if state {
		self.removeState()
	} else {
		self.restoreState()
	}

	return state
}

//
// Пропускает пробелы и возвращает кол-во пропущенных
//
func (self *parser) skipSpaces() int {
	count := 0
	for (self.curCharClass & SPACE) != NULL {
		self.moveNextPos()
		count++
	}
	return count
}

//
// Пропускает символы перевода строк и возвращает кол-во пропущенных символов
//
// limit int - лимит пропусков символов перевода строк, при установке в -1 - не лимитируется
//
func (self *parser) skipNL(limit int) int {
	count := 0
	for (self.curCharClass & NL) != NULL {
		if limit >= 0 && count == limit {
			break
		}

		self.moveNextPos()
		self.skipSpaces()

		count++
	}

	return count
}

//
// Пропускает символы относящиеся к классу и возвращает кол-во пропущенных символов
//
// class int - класс для пропуска
//
func (self *parser) skipClass(class int) int {
	count := 0
	for (self.curCharClass & class) != NULL {
		self.moveNextPos()
		count++
	}
	return count
}

//
// Захватывает все последующие символы относящиеся к классу и возвращает строку из этих символов
//
// class int - класс для захвата
//
func (self *parser) grabCharClass(class int) string {
	buff := []rune{}
	for (self.curCharClass & class) != NULL {
		buff = append(buff, self.curChar)
		self.moveNextPos()
	}
	return string(buff)
}

//
// Захватывает все последующие символы НЕ относящиеся к классу и возвращает строку из этих символов
//
// class int - класс для остановки захвата
//
func (self *parser) grabNotCharClass(class int) string {
	buff := []rune{}
	for self.curCharClass != NULL && ((self.curCharClass & class) == NULL) {
		buff = append(buff, self.curChar)
		self.moveNextPos()
	}
	return string(buff)
}

//
// Готовит контент, возвращает строку с готовым текстом
//
// parentTag string - имя родительского тега или пустая строка
//
func (self *parser) makeContent(parentTag string) string {
	content := bytes.NewBufferString("")

	self.skipSpaces()
	self.skipNL(-1)

	for self.curCharClass != NULL {
		tagName := ""
		tagParams := make(map[string]string)
		tagContent := ""
		shortTag := false

		// Если текущий тег это тег без текста, то пропускаем символы до "<"
		if _, ok := self.tagParentOnly[self.curTag]; ok && self.curChar != '<' {
			self.skipTextToChar('<')
		}

		self.saveState()

		switch {
		// Тег в котором есть текст
		case self.curChar == '<' && self.matchTag(&tagName, &tagParams, &tagContent, &shortTag):
			tagBuilt := self.makeTag(tagName, tagParams, tagContent, shortTag, parentTag)
			content.WriteString(tagBuilt)
			if _, ok := self.tagBlockType[tagName]; (ok || tagName == "br") && tagBuilt != "" {
				self.skipNL(1)
			}
			if tagBuilt == "" {
				self.skipClass(SPACE | NL)
			}
		// Комментарий <!-- -->
		case self.curChar == '<' && self.matchStr("<!--"):
			if self.skipTextToStr("-->") {
				self.skipStr("-->")
				self.skipClass(SPACE | NL)
			}
		// Конец тега
		case self.curChar == '<' && self.matchTagClose(&tagName):
			if self.curTag != "" {
				self.restoreState()
				return content.String()
			} else {
				self.setError(errors.New("Не ожидалось закрывающего тега '" + tagName + "'"))
			}
		// Просто символ "<"
		case self.curChar == '<':
			if _, ok := self.tagParentOnly[self.curTag]; !ok {
				content.WriteString(self.entities['<'])
			}
			self.moveNextPos()
		// Вероятно тут просто текст, формируем его
		default:
			content.WriteString(self.makeText(parentTag))
		}
		self.removeState()
	}

	return content.String()
}

//
// Обработка тега
//
// tagName *string - имя тега
// tagParams *map[string]string - параметры тега
// tagContent *string - контент тега
// shortTag *bool - короткий ли тег
//
func (self *parser) matchTag(tagName *string, tagParams *map[string]string, tagContent *string, shortTag *bool) bool {
	*tagName = ""
	*tagParams = make(map[string]string)
	*tagContent = ""
	*shortTag = false

	closeTag := ""

	if !self.matchTagOpen(tagName, tagParams, shortTag) {
		return false
	}

	if *shortTag {
		return true
	}

	curTag := self.curTag
	isTypoMode := self.isTypoMode

	if _, ok := self.tagNoTypography[*tagName]; ok {
		self.isTypoMode = false
	}

	self.curTag = *tagName

	if _, ok := self.tagPreformatted[*tagName]; ok {
		*tagContent = self.makePreformatted(*tagName)
	} else {
		*tagContent = self.makeContent(*tagName)
	}

	if self.matchTagClose(&closeTag) && *tagName != closeTag {
		self.setError(errors.New("Неверный закрывающийся тег '" + closeTag + "'. Ожидалось закрытие '" + *tagName + "'"))
	}

	self.curTag = curTag
	self.isTypoMode = isTypoMode

	return true
}

//
// Обработка открывающего тега
//
// tagName *string - имя тега
// tagParams *map[string]string - параметры тега
// shortTag *bool - короткий ли тег
//
func (self *parser) matchTagOpen(tagName *string, tagParams *map[string]string, shortTag *bool) bool {
	if self.curChar != '<' {
		return false
	}

	self.saveState()

	if self.skipSpaces() == 0 {
		self.moveNextPos()
	}

	*tagName = self.grabCharClass(TAG_NAME)

	self.skipSpaces()

	if *tagName == "" {
		self.restoreState()
		return false
	}

	*tagName = strings.ToLower(*tagName)

	if self.curChar != '>' && self.curChar != '/' {
		self.matchTagParams(tagParams)
	}

	_, *shortTag = self.tagShort[*tagName]

	if !*shortTag && self.curChar == '/' {
		self.restoreState()
		return false
	}

	if *shortTag && self.curChar == '/' {
		self.moveNextPos()
	}

	self.skipSpaces()

	if self.curChar != '>' {
		self.restoreState()
		return false
	}

	self.removeState()
	self.moveNextPos()

	return true
}

//
// Обработка закрывающего тега
//
// tagName *string - имя тега
//
func (self *parser) matchTagClose(tagName *string) bool {
	if self.curChar != '<' {
		return false
	}

	self.saveState()

	if self.skipSpaces() == 0 {
		self.moveNextPos()
	}

	if self.curChar != '/' {
		self.restoreState()
		return false
	}

	if self.skipSpaces() == 0 {
		self.moveNextPos()
	}

	*tagName = self.grabCharClass(TAG_NAME)

	self.skipSpaces()

	if *tagName == "" || self.curChar != '>' {
		self.restoreState()
		return false
	}

	*tagName = strings.ToLower(*tagName)

	self.removeState()
	self.moveNextPos()

	return true
}

//
// Обработка параметров тега
//
// params *map[string]string - карта параметров
//
//
func (self *parser) matchTagParams(params *map[string]string) bool {
	name := ""
	value := ""
	for self.matchTagParam(&name, &value) {
		if []rune(name)[0] != '-' {
			(*params)[name] = value
		}
		name, value = "", ""
	}
	return (len(*params) > 0)
}

//
// Обработка одного параметра тега
//
// name *string - имя параметра
// value *string - значение параметра
//
func (self *parser) matchTagParam(name *string, value *string) bool {
	self.saveState()
	self.skipSpaces()

	*name = self.grabCharClass(TAG_PARAM_NAME)

	if *name == "" {
		self.removeState()
		return false
	}

	self.skipSpaces()

	// Параметр без значения
	if self.curChar != '=' {
		if self.curChar == '>' || self.curChar == '/' || ((self.curCharClass & SPACE) != NULL) {
			*value = *name
			self.removeState()
			return true
		} else {
			self.restoreState()
			return false
		}
	} else {
		self.moveNextPos()
	}

	self.skipSpaces()

	if !self.matchTagParamValue(value) {
		self.restoreState()
		return false
	}

	self.skipSpaces()
	self.removeState()

	return true
}

//
// Обработка значения параметра тега
//
// value *string - значение параметра
//
func (self *parser) matchTagParamValue(value *string) bool {
	if (self.curCharClass & TAG_QUOTE) != NULL {
		quote := self.curChar
		escape := false

		self.moveNextPos()

		buff := bytes.NewBufferString("")
		for self.curCharClass != NULL && (self.curChar != quote || escape == true) {
			if _, ok := self.entities[self.curChar]; ok {
				buff.WriteString(self.entities[self.curChar])
			} else {
				buff.WriteString(string(self.curChar))
			}

			// Возможны экранированные кавычки
			escape = (self.curChar == '\\')

			self.moveNextPos()
		}

		*value = buff.String()

		if self.curChar != quote {
			return false
		}

		self.moveNextPos()
	} else {
		buff := bytes.NewBufferString("")
		for self.curCharClass != NULL && ((self.curCharClass & SPACE) == NULL) && self.curChar != '>' {
			if _, ok := self.entities[self.curChar]; ok {
				buff.WriteString(self.entities[self.curChar])
			} else {
				buff.WriteString(string(self.curChar))
			}
			self.moveNextPos()
		}

		*value = buff.String()
	}

	return true
}

//
// Готовит преформатированный контент
//
// openTag string - текущий открывающий тег
//
func (self *parser) makePreformatted(openTag string) string {
	content := bytes.NewBufferString("")
	for self.curCharClass != NULL {
		if self.curChar == '<' && openTag != "" {
			closeTag := ""
			self.saveState()

			isClosedTag := self.matchTagClose(&closeTag)

			if isClosedTag {
				self.restoreState()
			} else {
				self.removeState()
			}

			if isClosedTag && openTag == closeTag {
				break
			}
		}

		if _, ok := self.entities[self.curChar]; ok {
			content.WriteString(self.entities[self.curChar])
		} else {
			content.WriteString(string(self.curChar))
		}

		self.moveNextPos()
	}

	return content.String()
}

//
// Готовит тег к печати
//
// tagName string - имя тега
// tagParams map[string]string - параметры тега
// tagContent string - контент тега
// shortTag bool - короткий ли тег
// parentTag string - имя тега родителя, если есть
//
func (self *parser) makeTag(tagName string, tagParams map[string]string, tagContent string, shortTag bool, parentTag string) string {
	tagName = strings.ToLower(tagName)

	// Тег необходимо вырезать вместе с содержимым
	if _, ok := self.tagCutWithContent[tagName]; ok {
		return ""
	}

	// Допустим ли тег к использованию
	if _, ok := self.tagAllowed[tagName]; !ok {
		if _, ok := self.tagParentOnly[parentTag]; ok {
			return ""
		} else {
			return tagContent
		}
	}

	// Должен ли тег НЕ быть дочерним к любому другому тегу
	if _, ok := self.tagGlobalOnly[tagName]; ok && parentTag != "" {
		return tagContent
	}

	// Может ли тег находиться внутри родительского тега
	if _, ok := self.tagParentOnly[parentTag]; ok {
		if _, ok := self.tagChild[parentTag][tagName]; !ok {
			return ""
		}
	}

	// Тег может находиться только внутри другого тега
	if _, ok := self.tagChildOnly[tagName]; ok {
		if _, ok := self.tagParent[tagName][parentTag]; !ok {
			return tagContent
		}
	}

	// Параметры тега
	tagParamsResult := make(map[string]string)
	for param, value := range tagParams {
		param = strings.ToLower(param)
		value = strings.TrimSpace(value)

		if value == "" {
			continue
		}

		// Разрешен ли этот атрибут
		paramAllowedValues, isParamAllowedValues := self.tagParamAllowed[tagName][param]

		if !isParamAllowedValues {
			continue
		}

		found := false
		for _, paramAllowedValue := range paramAllowedValues {
			if paramAllowedValue == "#str" {
				found = true
				break
			} else if paramAllowedValue == "#int" {
				if matched, _ := regexp.MatchString(`^[0-9]+$`, value); matched {
					continue
				}
				found = true
				break
			} else if paramAllowedValue == "#link" {
				if matched, _ := regexp.MatchString(`javascript:`, value); matched {
					continue
				}

				if matched, _ := regexp.MatchString(`^(?i)[a-z0-9/#]`, value); !matched {
					continue
				}

				protocols := strings.Join(self.linkProtocolAllow, "|")
				if matched, _ := regexp.MatchString(`^(`+protocols+`):\/\/`, value); !matched {
					if matched, _ := regexp.MatchString(`^(\/|\#)`, value); !matched {
						value = "http://" + value
					}
				}

				found = true
				break
			} else if strings.HasPrefix(paramAllowedValue, "#regexp") {
				rx := regexp.MustCompile(`^#regexp\((.*?)\)$`)
				mc := rx.FindStringSubmatch(paramAllowedValue)

				if len(mc) < 2 {
					continue
				}

				if matched, _ := regexp.MatchString(mc[1], value); !matched {
					continue
				}

				found = true
				break
			} else if paramAllowedValue == value {
				found = true
				break
			}
		}

		if !found {
			self.setError(errors.New("Недопустимое значение '" + value + "' для атрибута '" + param + "' тега '" + tagName + "'"))
			continue
		}

		tagParamsResult[param] = value
	}

	// Проверка обязательных параметров тега
	if _, ok := self.tagParamRequired[tagName]; ok {
		for param := range self.tagParamRequired[tagName] {
			if _, ok := tagParamsResult[param]; !ok {
				return tagContent
			}
		}
	}

	// Авто добавляемые параметры
	if _, ok := self.tagParamDefault[tagName]; ok {
		for param, value := range self.tagParamDefault[tagName] {
			if _, ok := tagParamsResult[param]; !ok {
				tagParamsResult[param] = value
			}
		}
	}

	// Параметры значения которых должны быть именно такими
	if _, ok := self.tagParamReview[tagName]; ok {
		for param, value := range self.tagParamReview[tagName] {
			tagParamsResult[param] = value
		}
	}

	// Удаляем пустые не короткие теги если не сказано другого
	if _, ok := self.tagEmpty[tagName]; !ok {
		if !shortTag && tagContent == "" {
			return ""
		}
	}

	// Вызываем callback функцию, если тег собирается именно так
	if cb, ok := self.tagBuildCallback[tagName]; ok {
		return cb(tagName, tagParamsResult, tagContent)
	}

	// Собираем тег
	buff := bytes.NewBufferString("<" + tagName)

	for _, param := range self.tagParamSorted[tagName] {
		if value, ok := tagParamsResult[param]; ok {
			buff.WriteString(" " + param + "=\"" + value + "\"")
		}
	}

	if shortTag && self.isXHTMLMode {
		buff.WriteString("/>")
	} else {
		buff.WriteString(">")
	}

	if _, ok := self.tagParentOnly[tagName]; ok {
		buff.WriteString("\n")
	}

	if !shortTag {
		buff.WriteString(tagContent)
		buff.WriteString("</" + tagName + ">")
	}

	if _, ok := self.tagParentOnly[parentTag]; ok {
		buff.WriteString("\n")
	}

	if _, ok := self.tagBlockType[tagName]; ok {
		buff.WriteString("\n")
	}

	if tagName == "br" {
		buff.WriteString("\n")
	}

	return buff.String()
}

//
// Проверяет текущую позицию на вхождение тире пригодного для замены
//
// dash *string - тире
//
func (self *parser) matchDash(dash *string) bool {
	if self.curChar != '-' {
		return false
	}

	if ((self.prevCharClass & (SPACE | NL | TEXT_BRACKET)) == NULL) && self.prevCharClass != NULL {
		return false
	}

	self.saveState()

	for self.nextChar == '-' {
		self.moveNextPos()
	}

	if ((self.nextCharClass & (SPACE | NL | TEXT_BRACKET)) == NULL) && self.nextCharClass != NULL {
		self.restoreState()
		return false
	}

	*dash = self.dash
	self.removeState()
	self.moveNextPos()

	return true
}

//
// Определяет HTML сущности
//
// entity *string - HTML-сущность
//
func (self *parser) matchHTMLEntity(entity *string) bool {
	if self.curChar != '&' {
		return false
	}

	self.saveState()
	self.moveNextPos()

	if self.curChar == '#' {
		self.moveNextPos()

		entityCode := self.grabCharClass(NUMERIC)

		if entityCode == "" || self.curChar != ';' {
			self.restoreState()
			return false
		}

		self.removeState()
		self.moveNextPos()

		*entity = html.UnescapeString("&#" + entityCode + ";")

		return true
	} else {
		entityName := self.grabCharClass(ALPHA | NUMERIC)

		if entityName == "" || self.curChar != ';' {
			self.restoreState()
			return false
		}

		self.removeState()
		self.moveNextPos()

		*entity = html.UnescapeString("&" + entityName + ";")

		return true
	}
}

//
// Проверяет текущую позицию на вхождение кавычки пригодной для замены
//
// quote *string - кавычка
//
func (self *parser) matchQuote(quote *string) bool {
	if (self.curCharClass & TEXT_QUOTE) == NULL {
		return false
	}

	tp := "open"
	if (self.quotesOpened >= 2) ||
		((self.quotesOpened > 0) &&
			((((self.prevCharClass & (SPACE | NL | TEXT_BRACKET)) == NULL) && self.prevCharClass != NULL) ||
				(((self.nextCharClass & (SPACE | NL | TEXT_BRACKET | PUNCTUATUON)) != NULL) || self.nextCharClass == NULL))) {
		tp = "close"
	}

	if tp == "open" && ((self.prevCharClass & (SPACE | NL | TEXT_BRACKET)) == NULL) && self.prevCharClass != NULL {
		return false
	}

	if tp == "close" && (self.nextCharClass&(SPACE|NL|TEXT_BRACKET|PUNCTUATUON)) == NULL && self.nextCharClass != NULL {
		return false
	}

	level := 0
	index := 0

	if tp == "open" {
		self.quotesOpened += 1
		level = self.quotesOpened - 1
		index = 0
	} else {
		self.quotesOpened -= 1
		level = self.quotesOpened
		index = 1
	}

	*quote = string(self.quotes[level][index])

	self.moveNextPos()

	return true
}

//
// Формирование текста
//
// parentTag string - возможный родительский тег
//
func (self *parser) makeText(parentTag string) string {
	text := bytes.NewBufferString("")

	for self.curChar != '<' && self.curCharClass != NULL {
		brCount := 0
		spResult := ""
		entity := ""
		quote := ""
		dash := ""
		url := ""

		switch {
		// Преобразование HTML сущностей
		case self.curChar == '&' && self.matchHTMLEntity(&entity):
			if val, ok := self.entities[[]rune(entity)[0]]; ok {
				text.WriteString(val)
			} else {
				text.WriteString(entity)
			}
		// Добавление символов пунктуации
		case (self.curCharClass & PUNCTUATUON) != NULL:
			text.WriteRune(self.curChar)
			self.moveNextPos()
		// Преобразование символов тире в длинное тире
		case self.isTypoMode && self.curChar == '-' && self.matchDash(&dash):
			text.WriteString(dash)
		// Преобразование кавычек
		case self.isTypoMode && ((self.curCharClass & TEXT_QUOTE) != NULL) && self.matchQuote(&quote):
			text.WriteString(quote)
		// Преобразование пробельных символов
		case (self.curCharClass & SPACE) != NULL:
			self.skipSpaces()
			text.WriteString(" ")
		// Преобразование символов перевода строк в тег <br>
		case self.isAutoBrMode && ((self.curCharClass & NL) != NULL):
			brCount = self.skipNL(-1)
			if _, ok := self.tagNoAutoBr[self.curTag]; !ok {
				br := self.br + "\n"
				if brCount == 1 {
					text.WriteString(br)
				} else {
					text.WriteString(br + br)
				}
			}
		// Преобразование текста похожего на ссылку в кликабельную ссылку
		case self.isAutoLinkMode && ((self.curCharClass & ALPHA) != NULL) && self.curTag != "a" && self.matchURL(&url):
			text.WriteString(self.makeTag("a", map[string]string{"href": url}, url, false, parentTag))
		// Вызов callback-функции если строка предварена специальным символом
		case self.isSpecialCharMode && ((self.curCharClass & SPECIAL_CHAR) != NULL) && self.curTag != "a" && self.matchSpecialChar(&spResult):
			text.WriteString(spResult)
		// Другие печатные символы
		case ((self.curCharClass & PRINATABLE) != NULL):
			if val, ok := self.entities[self.curChar]; ok {
				text.WriteString(val)
			} else {
				text.WriteRune(self.curChar)
			}
			self.moveNextPos()
		// Не печатные символы
		default:
			self.moveNextPos()
		}
	}

	return text.String()
}

//
// Определяет текстовые ссылки
//
// url *string - ссылка
//
func (self *parser) matchURL(url *string) bool {
	if ((self.prevCharClass & (SPACE | NL | TEXT_QUOTE | TEXT_BRACKET)) == NULL) && self.prevCharClass != NULL {
		return false
	}

	self.saveState()

	switch {
	case self.matchStr("http://") && IndexStringSlice(self.linkProtocolAllow, "http") != -1:
		break
	case self.matchStr("https://") && IndexStringSlice(self.linkProtocolAllow, "https") != -1:
		break
	case self.matchStr("ftp://") && IndexStringSlice(self.linkProtocolAllow, "ftp") != -1:
		break
	case self.matchStr("www."):
		*url = "http://"
	default:
		self.restoreState()
		return false
	}

	openBracket := rune(0)
	if (self.prevCharClass & TEXT_BRACKET) != NULL {
		if _, ok := self.bracketsALL[self.prevChar]; ok {
			openBracket = self.prevChar
		}
	}

	closeBracket := rune(0)
	if openBracket != 0 {
		closeBracket, _ = self.bracketsALL[self.prevChar]
	}

	openedBracket := 0
	if openBracket != 0 {
		openedBracket = 1
	}

	buff := bytes.NewBufferString("")
	// Метка выхода из цикла для switch{..}
BREAKNOW:
	for (self.curCharClass & PRINATABLE) != NULL {
		switch {
		case self.curChar == '<':
			break BREAKNOW
		case (self.curCharClass & TEXT_QUOTE) != NULL:
			break BREAKNOW
		case ((self.curCharClass & TEXT_BRACKET) != NULL) && openedBracket > 0:
			if self.curChar == closeBracket && openedBracket == 1 {
				break BREAKNOW
			}
			if self.curChar == openBracket {
				openedBracket += 1
			}
			if self.curChar == closeBracket {
				openedBracket -= 1
			}
		case (self.curCharClass & PUNCTUATUON) != NULL:
			self.saveState()
			punctuatuon := self.grabCharClass(PUNCTUATUON)

			if (self.curCharClass & PRINATABLE) == NULL {
				self.restoreState()
				break BREAKNOW
			}

			self.removeState()
			buff.WriteString(punctuatuon)

			if (self.curCharClass & (TEXT_QUOTE | TEXT_BRACKET)) != NULL {
				break BREAKNOW
			}
		}

		buff.WriteRune(self.curChar)
		self.moveNextPos()
	}

	if buff.String() == "" {
		self.restoreState()
		return false
	}

	self.removeState()
	*url += buff.String()

	return true
}

//
// Определяет строки предваренные спецсимволами
//
// spResult *string - результат работы callback-функции
//
func (self *parser) matchSpecialChar(spResult *string) bool {
	if (self.curCharClass & SPECIAL_CHAR) == NULL {
		return false
	}

	if _, ok := self.specialChars[self.curChar]; !ok {
		return false
	}

	if self.prevCharClass != NULL && ((self.prevCharClass & (SPACE | NL | TEXT_BRACKET)) == NULL) {
		return false
	}

	buff := bytes.NewBufferString("")
	spChar := self.curChar

	self.saveState()
	self.moveNextPos()

	for self.curCharClass != NULL && ((self.curCharClass & (SPACE | NL | TEXT_BRACKET)) == NULL) {
		if (self.curCharClass & PUNCTUATUON) != NULL {
			self.saveState()

			punctuatuon := self.grabCharClass(PUNCTUATUON)

			if ((self.curCharClass & (SPACE | NL | TEXT_BRACKET)) != NULL) || self.curCharClass == NULL {
				self.restoreState()
				break
			}

			self.removeState()
			buff.WriteString(punctuatuon)
		}

		buff.WriteRune(self.curChar)
		self.moveNextPos()
	}

	if buff.String() == "" {
		self.restoreState()
		return false
	}

	*spResult = self.specialChars[spChar](buff.String())

	if *spResult == "" {
		self.restoreState()
		return false
	}

	self.removeState()

	return true
}

//
// Добавляет сообщение об ошибке
//
// msg error - сообщение об ошибке
//
func (self *parser) setError(msg error) {
	self.errorsList = append(self.errorsList, msg)
}

//
// Сравнение двух срезов рун
//
func EqualSliceRune(a, b []rune) bool {
	if len(a) != len(b) {
		return false
	}

	for i, c := range a {
		if c != b[i] {
			return false
		}
	}

	return true
}

//
// Поиск в элемента в срезе строк, возвращает индекс первого найденого или -1
//
func IndexStringSlice(s []string, e string) int {
	for i, c := range s {
		if c == e {
			return i
		}
	}
	return -1
}
