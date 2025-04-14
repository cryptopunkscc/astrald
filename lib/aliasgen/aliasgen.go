package aliasgen

import (
	"math/rand"
	"time"
)

var adjectives = []string{
	"zippy", "quirky", "snappy", "bouncy", "fizzy", "jolly", "wacky", "spunky", "zesty", "nifty",
	"breezy", "funky", "glitzy", "perky", "sassy", "cheery", "groovy", "lively", "peppy", "dandy",
	"kooky", "spry", "giddy", "plucky", "zany", "crisp", "whizzy", "jaunty", "merry", "spiffy",
	"brisk", "sunny", "flashy", "bubbly", "chirpy", "zippo", "nutty", "glowy", "frisky", "snazzy",
	"tingly", "jumpy", "vroomy", "posh", "dazzly", "munchy", "twirly", "glinty", "swoopy", "boomy",
	"fluffy", "grumpy", "shiny", "wobbly", "curly", "bumpy", "furry", "rusty", "silky", "puffy",
	"dizzy", "huffy", "misty", "goofy", "tasty", "chunky", "slimy", "frothy", "crunchy", "sparkly",
	"wiggly", "toasty", "jiggly", "mellow", "zappy", "loopy", "tangy", "snugly", "whacky", "buzzy",
	"picky", "yummy", "spongy", "gleamy", "tickle", "flinky", "chubby", "slinky", "roly", "blinky",
	"swirly", "pumpy", "gushy", "flappy", "skimpy", "hazy", "bloop", "clinky", "scampy", "twinky",
}

var nouns = []string{
	"widget", "gizmo", "doodle", "thingy", "whiz", "gadget", "blip", "spark", "zip", "bolt",
	"flick", "pop", "beep", "whirl", "jolt", "clank", "fuzz", "glint", "ping", "zest",
	"dash", "bloom", "chirp", "flare", "glimmer", "hoot", "jangle", "klick", "loop", "munch",
	"nudge", "plop", "quip", "rumble", "scramble", "tingle", "vibe", "wobble", "yelp", "zoom",
	"bop", "clunk", "drip", "flop", "gloop", "hiss", "jiggle", "knack", "lurch", "mingle",
	"nip", "ooze", "prance", "quirk", "ripple", "skid", "thump", "ump", "vroom", "whack",
	"xylo", "yawn", "zinger", "bump", "crunch", "ding", "fling", "grunt", "hush", "jive",
	"kick", "lick", "mist", "nibble", "pounce", "roar", "sip", "tweak", "whirl", "yip",
	"zap", "boing", "clash", "drift", "fluff", "gush", "hop", "jerk", "kiss", "lash",
	"moan", "nook", "puff", "riff", "sizzle", "tock", "vex", "wink", "yell", "zonk",
}

// New returns a random name in the format "adjective-noun"
func New() string {
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))

	adj := adjectives[rnd.Intn(len(adjectives))]
	noun := nouns[rnd.Intn(len(nouns))]

	return adj + "-" + noun
}
