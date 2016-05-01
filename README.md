#How does a GoPane work?
A lot like tmux, but within a Go program. A GoPane lets you split it horizontally or vertically. Thus, a GoPane has either zero or two elements, depending on if it's split or not.
If it has zero, then it's a leaf pane, and it can be printed to directly.
If it has been split, it will have two children, a first and a second GoPane. Thus the GoPane structure is a complete binary tree.
A print to a parent pane will try to print to the first element of each child until it reaches the
a leaf.
All GoPanes have a width and height which content will wrap to fit inside.
The width and height of the topmost parent (referred to as "root" in the source) will be the width and height of the window, retreived from the termbox-go library.

#What's supported?
- All terminals that github.com/nsf/termbox-go supports
- text wrapping, because we don't want to lose data
- vertical overflow (will scroll up and down)
    * TODO this is a later feature
- responsive resize detection (using polling)
    * TODO this is a later feature
    * when resized, panes will be resized by percentages
        - TODO unless otherwise specified?
- custom coloring (including dividers)
    * TODO this is yet another later feature

#What's not supported?
- padding or margins or any such styles - this is not a window manager
- boxes or close buttons or images

#Why does this exist?
- I don't like the existing solutions - I don't want a window manager or anything heavyweight
- I think tmux is incredibly useful and it'd be great to have inside my Go projects
- The terminal is great and this makes using a minimalistic terminal UI in Go very easy
- I think Go is fun and I like writing things in it, even if this isn't 100% necessary

##This is in a very pre-alpha (pretty much toy) state right now, definitely don't use it for anything that could touch production content.
