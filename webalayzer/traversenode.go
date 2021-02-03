package webalayzer

import (
	"context"
	"golang.org/x/net/html"
)

func traverseNode(ctx context.Context, node *html.Node, f func(context.Context, *html.Node) error) error {
	if node == nil {
		return nil
	}

	if err := f(ctx, node); err != nil {
		return err
	}

	if err := traverseNode(ctx, node.FirstChild, f); err != nil {
		return err
	}
	if err := traverseNode(ctx, node.NextSibling, f); err != nil {
		return err
	}
	return nil
}
