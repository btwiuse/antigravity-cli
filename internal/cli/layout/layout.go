package layout

type Dimensions struct {
	Width  int
	Height int
}

type Section struct {
	X      int
	Y      int
	Width  int
	Height int
}

type Shell struct {
	Header  Section
	Sidebar Section
	Body    Section
	Footer  Section
}

func Compute(d Dimensions) Shell {
	if d.Width <= 0 {
		d.Width = 120
	}
	if d.Height <= 0 {
		d.Height = 40
	}

	headerHeight := 3
	footerHeight := 2
	bodyHeight := d.Height - headerHeight - footerHeight
	if bodyHeight < 1 {
		bodyHeight = 1
	}

	sidebarWidth := 0
	if d.Width >= 100 {
		sidebarWidth = d.Width / 4
		if sidebarWidth < 24 {
			sidebarWidth = 24
		}
		if sidebarWidth > 32 {
			sidebarWidth = 32
		}
	}

	return Shell{
		Header:  Section{Width: d.Width, Height: headerHeight},
		Sidebar: Section{X: 0, Y: headerHeight, Width: sidebarWidth, Height: bodyHeight},
		Body:    Section{X: sidebarWidth, Y: headerHeight, Width: d.Width - sidebarWidth, Height: bodyHeight},
		Footer:  Section{Y: d.Height - footerHeight, Width: d.Width, Height: footerHeight},
	}
}
