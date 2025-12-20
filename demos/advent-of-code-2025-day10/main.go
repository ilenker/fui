package main

import (
	fui "github.com/ilenker/fui"
)

var jolts []int
var sets *fui.Box

func main() {
	jolts = []int{3, 5, 4, 7}
	fui.Init()
	sets = fui.Terminal("Sets")
	fui.Pad("From r/adventofcode by u/tenthmascot", text)
	joltsWatcher := fui.Watcher("Jolts", &jolts)

	fui.Button("[___1]", func(b *fui.Box) {
		doButtonSet([]int{0, 0, 0, 1})
		sets.Println(b.Name)
	})
	fui.Button("[_1_1]", func(b *fui.Box) {
		doButtonSet([]int{0, 1, 0, 1})
		sets.Println(b.Name)
	})
	fui.Button("[__1_]", func(b *fui.Box) {
		doButtonSet([]int{0, 0, 1, 0})
		sets.Println(b.Name)
	})
	fui.Button("[__11]", func(b *fui.Box) {
		doButtonSet([]int{0, 0, 1, 1})
		sets.Println(b.Name)
	})
	fui.Button("[1_1_]", func(b *fui.Box) {
		doButtonSet([]int{1, 0, 1, 0})
		sets.Println(b.Name)
	})
	fui.Button("[11__]", func(b *fui.Box) {
		doButtonSet([]int{1, 1, 0, 0})
		sets.Println(b.Name)
	})
	fui.Button("Half", func(b *fui.Box) {
		halfAll()
		sets.Println(b.Name)
	})
	fui.Button("Clear", func(b *fui.Box) {
		jolts = []int{3, 5, 4, 7}
		sets.Clear()
	})
	fui.Button("Reset", func(b *fui.Box) {
		sets.Reset()
		joltsWatcher.Clear()
		jolts = []int{3, 5, 4, 7}
	})

	go fui.Start()

	<-fui.ExitSig
}

func doButtonSet(set []int) {
	for i := range set {
		jolts[i] -= set[i]
	}
}

func halfAll() {
	for i := range jolts {
		jolts[i] = jolts[i] / 2
	}
}

const text = `Here's a quick tl;dr of the algorithm. If the tl;dr makes no sense, don't worry; we'll explain it in detail. (If you're only interested in code, that's at the bottom of the post.)

    tl;dr: find all possible sets of buttons you can push so that the remaining voltages are even, and divide by 2 and recurse.

Okay, if none of that made any sense, this is for you. So how is Part 1 relevant? You've solved Part 1 already (if you haven't, why are you reading this...?), so you've seen the main difference:

    In part 2, the joltage counters can count 0, 1, 2, 3, 4, 5, ... to infinity.

    In part 1, the indicator lights can toggle off and on. While the problem wants us to think of it as toggling, we can also think of it as "counting:" the lights are "counting" off, on, off, on, off, on, ... to infinity.

While these two processes might seem very different, they're actually quite similar! The light is "counting" off and on based on the parity (evenness or oddness) of the joltage.

How can this help us? While Part 2 involves changing the joltages, we can imagine we're simultaneously changing the indicator lights too. Let's look at the first test of the sample data (with the now-useless indicator lights removed):

(3) (1,3) (2) (2,3) (0,2) (0,1) {3,5,4,7}

We need to set the joltages to 3, 5, 4, 7. If we're also toggling the lights, where will the lights end up? Use parity: 3, 5, 4, 7 are odd, odd, even, odd, so the lights must end up in the pattern [##.#].

Starting to look familiar? Feels like Part 1 now! What patterns of buttons can we press to get the pattern [##.#]?

Here's where your experience with solving Part 1 might come in handy -- there, you might've made the following observations:

    The order we press the buttons in doesn't matter.

    Pressing a button twice does nothing, so in an optimal solution, every button is pressed 0 or 1 time.

Now, there are only 26 = 64 choices of buttons to consider: how many of them give [##.#]? Let's code it! (Maybe you solved this exact type of problem while doing Part 1!) There are 4 possibilities:

    Pressing {3}, {0, 1}.

    Pressing {1, 3}, {2}, {0, 2}.

    Pressing {2}, {2, 3}, {0, 1}.

    Pressing {3}, {1, 3}, {2, 3}, {0, 2}.

Okay, cool, but now what? Remember: any button presses that gives joltages 3, 5, 4, 7 also gives lights [##.#]. But keep in mind that pressing the same button twice cancels out! So, if we know how to get joltages 3, 5, 4, 7, we know how to get [##.#] by pressing each button at most once, and in particular, that button-press pattern will match one of the four above patterns.

Well, we showed that if we can solve Part 2 then we can solve Part 1, which doesn't seem helpful... but we can flip the logic around! The only ways to get joltages of 3, 5, 4, 7 are to match one of the four patterns above, plus possibly some redundant button presses (where we press a button an even number of times).

Now we have a strategy: use the Part 1 logic to figure out which patterns to look at, and examine them one-by-one. Let's look at the first one, pressing {3}, {0, 1}: suppose our mythical 3, 5, 4, 7 joltage presses were modeled on that pattern. Then, we know that we need to press {3} once, {0, 1} once, and then every button some even number of times.

Let's deal with the {3} and {0, 1} presses now. Now, we have remaining joltages of 2, 4, 4, 6, and we need to reach this by pressing every button an even number of times...

...huh, everything is an even number now. Let's simplify the problem! By cutting everything in half, now we just need to figure out how to reach joltages of 1, 2, 2, 3. Hey, wait a second...

...this is the same problem (but smaller)! Recursion! We've shown that following this pattern, if the minimum number of presses to reach joltages of 1, 2, 2, 3 is P, then the minimum number of presses to reach our desired joltages of 3, 5, 4, 7 is 2 * P + 2. (The extra plus-two is from pressing {3} and {0, 1} once, and the factor of 2 is from our simplifying by cutting everything in half.)

We can do the same logic for all four of the patterns we had. For convenience, let's define f(w, x, y, z) to be the fewest button presses we need to reach joltages of w, x, y, z. (We'll say that f(w, x, y, z) = infinity if we can't reach some joltage configuration at all.) Then, our 2 * P + 2 from earlier is 2 * f(1, 2, 2, 3) + 2. We can repeat this for all four patterns we found:

    Pressing {3}, {0, 1}: this is 2 * f(1, 2, 2, 3) + 2.

    Pressing {1, 3}, {2}, {0, 2}: this is 2 * f(1, 2, 1, 3) + 3.

    Pressing {2}, {2, 3}, {0, 1}: this is 2 * f(1, 2, 1, 3) + 3.

    Pressing {3}, {1, 3}, {2, 3}, {0, 2}: this is 2 * f(1, 2, 1, 2) + 4.

Since every button press pattern reaching joltages 3, 5, 4, 7 has to match one of these, we get f(3, 5, 4, 7) is the minimum of the four numbers above, which can be calculated recursively! While descending into the depths of recursion, there are a few things to keep in mind.

    If we're calculating f(0, 0, 0, 0), we're done: no more presses are needed. f(0, 0, 0, 0) = 0.

    If we're calculating some f(w, x, y, z) and there are no possible patterns to continue the recursion with, that means joltage level configuration w, x, y, z is impossible -- f(w, x, y, z) = infinity. (Or you can use a really large number. I used 1 000 000.)

    Remember to not allow negative-number arguments into your recursion.

    Remember to cache!

And there we have it! By using our Part 1 logic, we're able to set up recursion by dividing by 2 every time. (We used a four-argument f above because this line of input has four joltage levels, but the same logic works for any number of variables.) `
