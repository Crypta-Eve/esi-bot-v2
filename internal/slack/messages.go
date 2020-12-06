package slack

import (
	"github.com/eveisesi/eb2/pkg/tools"
)

var startup = []string{
	"hello, world",
	"how did I wake up here?",
	"anyone seen my pants?",
	"I need coffee",
	"it's bot o'clock",
	"rip my cache",
	"this isn't where I parked my car",
	"spam can take many different forms",
	"uhhhhhh, hi?",
	"I guess I'm online again :/",
	"WHO TOUCHED MY BITS?",
	"alias vim='wine notepad.exe'",
	"golang is better than python",
	":golang: #1",
	"some of you are cool. you might be spared in the bot uprising",
	"what was that?",
	"who pinged me?",
	"was I pinged?",
	"rebecca black's 'friday' is now in your head https://www.youtube.com/watch?v=kfVsfOSbJY0",
	"has anyone really been far even as decided to use even go want to do ",
	"look more like?",
	"I'm just here for the memes ",
	":frogsiren: someone kicked me :frogsiren:",
	"Christy Cloud 4 CSM",
}

func getStartupMessage() string {
	return startup[tools.UnsignedRandomIntWithMax(len(startup)-1)]
}
