package slack

import (
	"math/rand"
	"time"
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
	"vim > emacs",
	"rust is better than golang",
	":python:#1",
	"some of you are cool. you might be spared in the bot uprising",
	"what was that?",
	"who pinged me?",
	"was I pinged?",
	"rebecca black's 'friday' is now in your head",
	"has anyone really been far even as decided to use even go want to do ",
	"look more like?",
	"I'm just here for the memes",
	":frogsiren: someone kicked me :frogsiren:",
}

func getStartupMessage() string {
	return startup[getUnsignedRandomIntWithMax(len(startup)-1)]
}

func getUnsignedRandomIntWithMax(m int) uint {
	rand.Seed(time.Now().UnixNano())
	min := 0
	return uint(rand.Intn(m-min+1) + min)
}
