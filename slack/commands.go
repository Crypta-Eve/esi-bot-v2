package slack

import "strings"

// Links
const id string = "id"
const ids string = "ids"
const ranges string = "ranges"
const faq string = "faq"
const issues string = "issues"
const sso string = "sso"
const webui string = "webui"
const diff string = "diff"
const source string = "source"
const repo string = "repo"

var links = []string{faq,
	issues,
	sso,
	webui,
	diff,
	source,
	repo,
	id, ids, ranges,
}

var strLinks = strings.Join(links, ",")

// greetings
const hey string = "hey"
const hi string = "hi"
const hello string = "hello"
const ohseven string = "o7"
const sevenoh string = "7o"
const rightwave string = "o/"
const leftwave string = "\\o"

var greetings = []string{
	hey,
	hi,
	hello,
	ohseven,
	sevenoh,
	rightwave,
	leftwave,
}
var strGreetings = strings.Join(greetings, ",")

var commands = []string{
	strLinks, strGreetings,
}
