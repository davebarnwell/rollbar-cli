package ui

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"

	"rollbar-cli/internal/rollbar"
)

type UserRenderOptions struct {
	Fields    []string
	NoHeaders bool
}

func RenderUser(user rollbar.User) error {
	return renderUser(os.Stdout, user)
}

func RenderUsers(users []rollbar.User) error {
	return RenderUsersWithOptions(users, UserRenderOptions{})
}

func RenderUsersWithOptions(users []rollbar.User, opts UserRenderOptions) error {
	if len(users) == 0 {
		_, err := fmt.Fprintln(os.Stdout, "No users found.")
		return err
	}
	return renderUsersPlain(os.Stdout, users, opts)
}

func renderUser(w io.Writer, user rollbar.User) error {
	if _, err := fmt.Fprintf(w, "ID: %d\n", user.ID); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "Username: %s\n", fallback(user.Username)); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "Email: %s\n", fallback(user.Email)); err != nil {
		return err
	}
	return nil
}

func renderUsersPlain(w io.Writer, users []rollbar.User, opts UserRenderOptions) error {
	fields := opts.Fields
	if len(fields) == 0 {
		fields = []string{"id", "username", "email"}
	}
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	if !opts.NoHeaders {
		if _, err := fmt.Fprintln(tw, strings.Join(fieldHeaders(fields), "\t")); err != nil {
			return err
		}
	}
	for _, user := range users {
		if _, err := fmt.Fprintln(tw, strings.Join(userFieldValues(user, fields), "\t")); err != nil {
			return err
		}
	}
	return tw.Flush()
}

func userFieldValues(user rollbar.User, fields []string) []string {
	values := make([]string, 0, len(fields))
	for _, field := range fields {
		switch field {
		case "id":
			values = append(values, strconv.FormatInt(user.ID, 10))
		case "username":
			values = append(values, fallback(user.Username))
		case "email":
			values = append(values, fallback(user.Email))
		default:
			values = append(values, "-")
		}
	}
	return values
}
