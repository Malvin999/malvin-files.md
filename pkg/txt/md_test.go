package txt

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMDtoHTML(t *testing.T) {
	r := require.New(t)

	md := "### Header\n**bold**\n*italic*"
	html := MDtoHTML(md)

	r.Equal("<b>Header</b>\n<b>bold</b>\n<i>italic</i>", html)
}

func TestMDtoHTMLHeader(t *testing.T) {
	r := require.New(t)

	md := `# Header`
	html := MDtoHTML(md)

	r.Equal("<b>Header</b>", html)
}

func TestMDtoHTMLHeaderAndText(t *testing.T) {
	r := require.New(t)

	md := "# Header\nText"
	html := MDtoHTML(md)

	r.Equal("<b>Header</b>\nText", html)
}

func TestMDtoHTMlBold(t *testing.T) {
	r := require.New(t)

	md := "**bold**"
	html := MDtoHTML(md)

	r.Equal("<b>bold</b>", html)
}

func TestMDtoHtmlMultilineBold(t *testing.T) {
	r := require.New(t)

	md := "**bold\nstill bold**"
	html := MDtoHTML(md)

	r.Equal("<b>bold\nstill bold</b>", html)
}

func TestMDtoHTMLEmptyBoldAndItalic(t *testing.T) {
	r := require.New(t)

	md := "**"
	r.Equal("**", MDtoHTML(md))
	md = "*"
	r.Equal("*", MDtoHTML(md))

	md = "__"
	r.Equal("__", MDtoHTML(md))
	md = "_"
	r.Equal("_", MDtoHTML(md))
}

func TestMDtoHTMLNewLineChar(t *testing.T) {
	r := require.New(t)

	bold := "**\n**"
	r.Equal("<b>\n</b>", MDtoHTML(bold))

	italic := "*\n*"
	r.Equal("<i>\n</i>", MDtoHTML(italic))
}

func TestMDtoHTMLCharAndNewLineChar(t *testing.T) {
	r := require.New(t)

	bold := "**a\n**"
	r.Equal("<b>a\n</b>", MDtoHTML(bold))

	italic := "*a\n*"
	r.Equal("<i>a\n</i>", MDtoHTML(italic))

}

func TestMDtoHTMLNewLineAndChar(t *testing.T) {
	r := require.New(t)

	bold := "**\na**"
	r.Equal("<b>\na</b>", MDtoHTML(bold))

	italic := "*\na*"
	r.Equal("<i>\na</i>", MDtoHTML(italic))
}

func TestMDtoHTMLTwoNewlinesBreakFormatting(t *testing.T) {
	r := require.New(t)

	bold := "**no bold\n\nno bold**"
	r.Equal("**no bold\n\nno bold**", MDtoHTML(bold))

	italic := "*no italic\n\nno italic*"
	r.Equal("*no italic\n\nno italic*", MDtoHTML(italic))
}

func TestMDtoHTMlMultilineBoldAndItalic(t *testing.T) {
	r := require.New(t)

	md := "Some _italic text\nin two lines_, **bold text\nin two lines**, and ***bold italic text\nin two lines***."
	html := MDtoHTML(md)

	r.Equal("Some <i>italic text\nin two lines</i>, <b>bold text\nin two lines</b>, and <b><i>bold italic text\nin two lines</i></b>.", html)
}

func TestMDtoHTMLHtmlInsideCode(t *testing.T) {
	r := require.New(t)

	md := "```some code a > b```"
	html := MDtoHTML(md)

	r.Equal("<pre>some code a &gt; b</pre>", html)
}

func TestMDtoHTMLInlineCodeAndCodeBlock(t *testing.T) {
	r := require.New(t)

	md := "`inline code` and ```code block```"
	html := MDtoHTML(md)

	r.Equal("<code>inline code</code> and <pre>code block</pre>", html)
}

func TestMDtoHTMLInlineCodeBlockInsideCodeBlock(t *testing.T) {
	r := require.New(t)

	md := "```code block with `inline code` inside```"
	html := MDtoHTML(md)

	r.Equal("<pre>code block with <code>inline code</code> inside</pre>", html)
}

func TestMDToHTMLItalic(t *testing.T) {
	r := require.New(t)

	md := "*italic*"
	html := MDtoHTML(md)

	r.Equal("<i>italic</i>", html)
}

func TestMDToHTMLInvalidMD(t *testing.T) {
	r := require.New(t)

	md := "__valid__**invalid"
	r.Equal("<b>valid</b>**invalid", MDtoHTML(md))

	r.Equal("*invalid_markdown", MDtoHTML("*invalid_markdown"))
	r.Equal("*``invalid_markdown", MDtoHTML("*``invalid_markdown"))
	r.Equal("*```invalid_markdown", MDtoHTML("*```invalid_markdown"))
}

func TestMDToHTMLMultiline(t *testing.T) {
	r := require.New(t)

	md := "line1\n**line2**\nline3"
	html := MDtoHTML(md)

	r.Equal("line1\n<b>line2</b>\nline3", html)
}

func TestMDToHTMLBoldInsideItalic(t *testing.T) {
	r := require.New(t)

	md := "*italic and __bold__*"
	r.Equal("<i>italic and <b>bold</b></i>", MDtoHTML(md))

	md = "*italic and **bold***"
	r.Equal("<i>italic and <b>bold</b></i>", MDtoHTML(md))
}

func TestMDToHTMLItalicInsideBold(t *testing.T) {
	r := require.New(t)

	md := "__bold and _italic___"
	r.Equal("<b>bold and <i>italic</i></b>", MDtoHTML(md))

	md = "**bold and *italic***"
	r.Equal("<b>bold and <i>italic</i></b>", MDtoHTML(md))
}

func TestMDtoHTMLNoLists(t *testing.T) {
	r := require.New(t)

	md := "list\n1) item1\n2) item2"
	html := MDtoHTML(md)

	r.Equal("list\n1) item1\n2) item2", html)
}

func TestMDToHTMLEscapeHtml(t *testing.T) {
	r := require.New(t)

	html := MDtoHTML("<a> &b")

	r.Equal("&lt;a&gt; &amp;b", html)
}

func TestMDToHTMLHeader(t *testing.T) {
	r := require.New(t)

	md := "Multiline\n# Header"
	html := MDtoHTML(md)

	r.Equal("Multiline\n<b>Header</b>", html)
}

func TestMDtoHTMLMultipleHeaders(t *testing.T) {
	r := require.New(t)

	md := "# Header1\n## Header2"
	html := MDtoHTML(md)

	r.Equal("<b>Header1</b>\n<b>Header2</b>", html)
}

func TestMDtoHTMLInlineCode(t *testing.T) {
	r := require.New(t)

	md := "`inline code`"
	html := MDtoHTML(md)

	r.Equal("<code>inline code</code>", html)
}

func TestMDtoHTMLMultilineCodeBlock(t *testing.T) {
	r := require.New(t)

	md := "```\ncode line 1\ncode line 2\n```"
	html := MDtoHTML(md)

	r.Equal("<pre>\ncode line 1\ncode line 2\n</pre>", html)
}

func TestMDtoHTMLCodeWithBold(t *testing.T) {
	r := require.New(t)

	md := "`code` **bold**"
	html := MDtoHTML(md)

	r.Equal("<code>code</code> <b>bold</b>", html)
}

func TestMDtoHTMLHeaderWithInlineCode(t *testing.T) {
	r := require.New(t)

	md := "# Header\n`inline code`"
	html := MDtoHTML(md)

	r.Equal("<b>Header</b>\n<code>inline code</code>", html)
}
