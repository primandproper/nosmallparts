package components

import (
	"fmt"
	"log"
	"strings"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/cmd/services/admin/design"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	g "maragu.dev/gomponents"
	ghtml "maragu.dev/gomponents/html"
)

type ComponentRenderer struct {
	palette design.Palette
}

func NewComponentRenderer() *ComponentRenderer {
	return &ComponentRenderer{palette: design.StandardPalette}
}

func (r *ComponentRenderer) UsernameInput(label, fieldName, content string) g.Node {
	return ghtml.Input(
		ghtml.Type("text"),
		ghtml.ID(label),
		ghtml.Name(fieldName),
		ghtml.Value(content),
		ghtml.Class(fmt.Sprintf("mt-1 block w-full rounded-md border-%s shadow-sm focus:border-%s focus:ring-%s", r.palette.Background.Value, r.palette.Primary.Value, r.palette.Primary.Value)),
		ghtml.AutoComplete("username"),
	)
}

func (r *ComponentRenderer) passwordInput(id, fieldName, content string) g.Node {
	return ghtml.Input(
		ghtml.Type("password"),
		ghtml.ID(id),
		ghtml.Name(fieldName),
		ghtml.Value(content),
		ghtml.Class(fmt.Sprintf("mt-1 block w-full rounded-md border-%s shadow-sm focus:border-%s focus:ring-%s", r.palette.Background.Value, r.palette.Primary.Value, r.palette.Primary.Value)),
		ghtml.AutoComplete("current-password"),
	)
}

func (r *ComponentRenderer) totpTokenInput(id, fieldName, content string) g.Node {
	return ghtml.Input(
		ghtml.Type("text"),
		ghtml.ID(id),
		ghtml.Name(fieldName),
		ghtml.Value(content),
		ghtml.Class(fmt.Sprintf("mt-1 block w-full rounded-md border-%s shadow-sm focus:border-%s focus:ring-%s", r.palette.Background.Value, r.palette.Primary.Value, r.palette.Primary.Value)),
		ghtml.MaxLength("6"),
		g.Attr("inputmode", "numeric"),
		ghtml.Pattern("[0-9]{6}"),
		ghtml.AutoComplete("one-time-code"),
	)
}

func (r *ComponentRenderer) wrapInputElement(
	label,
	inputError string,
	input g.Node,
) g.Node {
	titleCaser := cases.Title(language.English)

	return ghtml.Div(
		ghtml.Class("space-y-1"),
		ghtml.Label(
			ghtml.For(label),
			ghtml.Class(fmt.Sprintf("block text-sm font-medium %s", design.TextColor(r.palette.Primary))),
			g.Text(titleCaser.String(label)),
		),
		input,
		g.If(inputError != "", ghtml.Span(
			ghtml.Class(fmt.Sprintf("text-sm %s mt-1 block", design.TextColor(r.palette.Warning))),
			g.Text(inputError),
		)),
	)
}

func (r *ComponentRenderer) inputElement(
	label,
	inputError,
	inputType,
	fieldName,
	content string,
) g.Node {
	var input g.Node
	s := strings.ToLower(strings.TrimSpace(inputType))

	switch s {
	case "username":
		input = r.UsernameInput(label, fieldName, content)
	case "password":
		input = r.passwordInput("password", fieldName, content)
	case "totp":
		input = r.totpTokenInput("totp", fieldName, content)
	default:
		log.Panicf("unknown input type: %s\n", s)
	}

	titleCaser := cases.Title(language.English)

	return ghtml.Div(
		ghtml.Class("space-y-1"),
		ghtml.Label(
			ghtml.For(label),
			ghtml.Class(fmt.Sprintf("block text-sm font-medium %s", design.TextColor(r.palette.Primary))),
			g.Text(titleCaser.String(label)),
		),
		input,
		g.If(inputError != "", ghtml.Span(
			ghtml.Class(fmt.Sprintf("text-sm %s mt-1 block", design.TextColor(r.palette.Warning))),
			g.Text(inputError),
		)),
	)
}

func submitButton(text string) g.Node {
	return ghtml.Button(
		ghtml.Type("submit"),
		ghtml.Class("w-full py-2 px-4 bg-blue-600 text-white font-semibold rounded-md shadow hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:opacity-50"),
		g.Text(text),
	)
}

type LoginFormProps struct {
	UsernameError,
	PasswordError,
	TOTPError,
	GeneralError string
}

// LoginForm renders a login form. You can optionally provide error messages for each field and a general error.
// Pass empty strings to ignore them.
func (r *ComponentRenderer) LoginForm(props *LoginFormProps) g.Node {
	return ghtml.Div(
		ghtml.ID("login-container"),
		ghtml.Div(
			ghtml.Class("w-full max-w-md bg-white p-8 rounded-2xl shadow-md"),
			ghtml.H2(
				ghtml.Class(fmt.Sprintf("text-2xl font-bold mb-6 text-center %s", design.TextColor(r.palette.Primary))),
				g.Text("Login"),
			),

			ghtml.Form(
				ghtml.Class("space-y-4"),
				ghtml.Method("post"),
				g.Attr("hx-post", "/login/submit"),
				g.Attr("hx-ext", "json-enc"),
				g.Attr("hx-target", "#login-container"),
				g.Attr("hx-swap", "outerHTML"),

				r.wrapInputElement("username", props.UsernameError, r.UsernameInput("username", "username", "")),
				r.inputElement("password", props.PasswordError, "password", "password", ""),
				r.inputElement("TOTP code (optional)", props.TOTPError, "totp", "totpToken", ""),

				submitButton("Log In"),

				g.If(props.GeneralError != "", ghtml.Div(
					ghtml.Class(fmt.Sprintf("mt-2 text-sm %s", design.TextColor(r.palette.Warning))),
					g.Text(props.GeneralError),
				)),
			),
		),
	)
}
