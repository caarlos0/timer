#!/bin/sh
set -e
rm -rf completions
mkdir completions
go build -o timer .
for sh in bash zsh fish; do
	./timer completion "$sh" >"completions/timer.$sh"
done
