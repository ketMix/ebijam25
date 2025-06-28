package client

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/kettek/rebui"
	"github.com/kettek/rebui/widgets"
)

type Dialoggies struct {
	layout       rebui.Layout
	dialogs      []*Dialog
	titleNode    *rebui.Node
	messageNode  *rebui.Node
	buttonsNodes []*rebui.Node
}

type Dialog struct {
	Title          string
	runningMessage string
	Message        string
	Buttons        []string
	OnSubmit       func(string)
}

type DialogButton struct {
	widgets.Button
	OnClick func()
}

func (b *DialogButton) HandlePointerPressed(evt rebui.EventPointerPressed) {
	b.OnClick()
}

func (d *Dialoggies) Add(title, message string, buttons []string, onSubmit func(string)) {
	d.dialogs = append(d.dialogs, &Dialog{
		Title:    title,
		Message:  message,
		Buttons:  buttons,
		OnSubmit: onSubmit,
	})
	if len(d.dialogs) == 1 {
		d.Next()
		fmt.Println("NEXT")
	}
}

func (d *Dialoggies) SetTitleColor(c color.NRGBA) {
	if d.titleNode != nil {
		d.titleNode.Widget.(*widgets.Text).AssignBackgroundColor(c)
	}
}

func (d *Dialoggies) Update() {
	if len(d.dialogs) == 0 {
		return
	}

	top := d.dialogs[0]
	if len(top.runningMessage) >= len(top.Message) {
		/*if inpututil.IsKeyJustReleased(ebiten.KeyEnter) || inpututil.IsKeyJustReleased(ebiten.KeySpace) {
			// Remove the dialog from the stack.
			d.dialogs = d.dialogs[1:]
			d.Next()
		}*/
	} else {
		// Increment the running message by 1 char.
		top.runningMessage = string(top.Message[0 : len(top.runningMessage)+1])
		d.messageNode.Widget.(*widgets.Text).AssignText(top.runningMessage)
	}
	d.layout.Update()
}

func (d *Dialoggies) Next() {
	if len(d.dialogs) == 0 {
		return
	}

	top := d.dialogs[0]
	//d.dialogs = d.dialogs[1:]

	if len(d.dialogs) == 0 {
		d.layout.RemoveNode(d.titleNode)
		d.layout.RemoveNode(d.messageNode)
		for _, node := range d.buttonsNodes {
			d.layout.RemoveNode(node)
		}
		d.buttonsNodes = nil
		return
	} else {
		if d.titleNode == nil {
			d.titleNode = d.layout.AddNode(rebui.Node{
				Type:            "Text",
				ID:              "title",
				Width:           "60%",
				Height:          "30",
				X:               "50%",
				Y:               "25%",
				OriginX:         "-50%",
				OriginY:         "-50%",
				ForegroundColor: "white",
				BackgroundColor: "black",
				Text:            top.Title,
				VerticalAlign:   rebui.AlignMiddle,
				HorizontalAlign: rebui.AlignCenter,
			})
		} else {
			d.titleNode.Widget.(*widgets.Text).AssignText(top.Title)
		}
		top.runningMessage = ""
		if d.messageNode == nil {
			d.messageNode = d.layout.AddNode(rebui.Node{
				Type:            "Text",
				ID:              "message",
				X:               "at title",
				Y:               "after title",
				Width:           "60%",
				Height:          "210",
				ForegroundColor: "white",
				BackgroundColor: "#444444",
				TextWrap:        rebui.WrapWord,
				Text:            top.runningMessage,
				/*VerticalAlign:   rebui.AlignMiddle,
				HorizontalAlign: rebui.AlignCenter,*/
			})
		} else {
			d.messageNode.Widget.(*widgets.Text).AssignText(top.runningMessage)
		}
		if d.buttonsNodes == nil {
			for i, button := range top.Buttons {
				x := "at message"
				if i > 0 {
					x = "after button_" + fmt.Sprintf("%d", i-1)
				}
				node := d.layout.AddNode(rebui.Node{
					Type:            "DialogButton",
					ID:              "button_" + fmt.Sprintf("%d", i),
					Width:           "60%",
					Height:          "30",
					X:               x,
					Y:               "after message",
					ForegroundColor: "white",
					BackgroundColor: "black",
					Text:            button,
					VerticalAlign:   rebui.AlignMiddle,
					HorizontalAlign: rebui.AlignCenter,
					FocusIndex:      i + 2, // +2 because title and message take up 1 each
				})
				node.Widget.(*DialogButton).OnClick = func() {
					top.OnSubmit("button_" + fmt.Sprintf("%d", i))
				}
			}
		}
	}
}

func (d *Dialoggies) Draw(screen *ebiten.Image) {
	if len(d.dialogs) == 0 {
		return
	}
	d.layout.Draw(screen)
}

func (d *Dialoggies) Layout(ow, oh float64) {
	d.layout.Layout(ow, oh)
}

func init() {
	rebui.RegisterWidget("DialogButton", &DialogButton{})
}
