// Program window-toggle-for-i3 focuses a window (based on its name) on the current
// workspace or starts a new instance of it.
package main

import (
	"flag"
	"fmt"
	"log"
	"regexp"

	"go.i3wm.org/i3/v4"
)

func logic() error {
	var (
		titleExpr = flag.String(
			"title_regexp",
			"- Google Chrome$",
			"Go regular expression (https://golang.org/pkg/regexp) that will be matched on the window title")

		cmd = flag.String(
			"not_found_command",
			"exec google-chrome",
			"i3 command to run if no window matching -title_regexp is found")

		scope = flag.String(
			"scope",
			"workspace",
			"workspace or root, specifies which child windows to match")

		mark = flag.String(
			"mark",
			"__last__",
			"Name of the mark used on the window before switching to target window")
	)
	flag.Parse()

	titleRe, err := regexp.Compile(*titleExpr)
	if err != nil {
		return err
	}

	tree, err := i3.GetTree()
	if err != nil {
		return err
	}

	var parent *i3.Node
	if *scope == "workspace" {
		parent = tree.Root.FindFocused(func(n *i3.Node) bool { return n.Type == i3.WorkspaceNode })
		if parent == nil {
			return fmt.Errorf("could not locate workspace")
		}
	} else {
		parent = tree.Root
	}

	focused := parent.FindChild(func(n *i3.Node) bool { return n.Focused })
	if titleRe.MatchString(focused.Name) {

		// If the target window is already focused, switch back to the window we
		// were using before focusing the target window
		_, err = i3.RunCommand(fmt.Sprintf("[con_mark=%s] focus", *mark))
		return err

	} else {

		// Otherwise, mark the current window with a custom mark so we can switch
		// back to it later
		i3_cmd := fmt.Sprintf("mark %s", *mark)
		_, err = i3.RunCommand(i3_cmd)
		if err != nil {
			return err
		}

	}

	if chrome := parent.FindChild(func(n *i3.Node) bool { return titleRe.MatchString(n.Name) }); chrome != nil {
		_, err = i3.RunCommand(fmt.Sprintf(`[con_id="%d"] focus`, chrome.ID))
	} else {
		_, err = i3.RunCommand(*cmd)
	}

	return err
}

func main() {
	if err := logic(); err != nil {
		log.Fatal(err)
	}
}
