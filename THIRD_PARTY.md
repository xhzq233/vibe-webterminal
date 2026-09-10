# Third-party components

The Go gateway and browser integration are adapted from
[dark2momo/herdr-tty](https://github.com/dark2momo/herdr-tty), via our
[xhzq233/herdr-tty fork](https://github.com/xhzq233/herdr-tty) at commit
`dba09a2`. The upstream MIT license is retained in
[third_party/herdr-tty.LICENSE](third_party/herdr-tty.LICENSE).

Runtime dependencies, installed separately:

- [ttyd](https://github.com/tsl0922/ttyd): MIT; includes xterm.js and its frontend
  dependencies. This version uses stock ttyd and no local Basic-auth patch.
- [Herdr](https://herdr.dev): persistent terminal sessions and native tabs.

The floating panel uses native JavaScript and CSS. interact.js and the old
vendored ttyd build are no longer part of this distribution.
