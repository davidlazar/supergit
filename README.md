# Supergit

Supergit is a tool for managing several git repositories as one.
More than that, Supergit's philosophy is that every file should be part
of a git repository, so it helps you find *untracked* files. In a sense,
Supergit treats its input directory as one large git repo composed of
smaller repos.

## Example

```
$ supergit ~/go
! bin
! pkg
? src/github.com/davidlazar/6.857coin
  ?? blocks/
  ?? logs/

! src/github.com/davidlazar/encmail
? src/github.com/davidlazar/flycrypt
   M flycrypt.go
  ?? flycrypt

? src/vuvuzela.io/alpenhorn
   M pkg/data.go
   M pkg/data_test.go
   M pkg/register.go
   M pkg/verify.go
  ?? TODO.txt
  ?? benchmarks/
  ?? configs/
  ?? metrics/
  ?? scratch
  (2 more lines)

! src/vuvuzela.io/configs
? src/vuvuzela.io/crypto
  ?? rand/laplace_test.go

? src/vuvuzela.io/vuvuzela
   M cmd/vuvuzela-client/conversation.go
```

