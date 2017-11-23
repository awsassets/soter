# Soter - An IRC Bot, Protector of channels

From the Greek god [Soter](https://en.wikipedia.org/wiki/Soter_(daimon))

> Soter (Σωτήρ "Saviour, Deliverer") was the personification or daimon of
> safety, preservation and deliverance from harm.

And so `Soter` is an IRC Bot that preserves and protects IRC Channels.

## Requirements

The main requirement of Soter is simply that is has IRC Operator privileges
on the server/network is it used on. On most IRCD(s) (*IRC Server software*)
this is called an O-line. Please make sure it has one!

## Installation

```#!bash
$ go get github.com/prologic/soter
```

## Getting Started

Simply run `soter`:

```#!bash
$ ./soter
```

## How it works

- Soter will connect to a configured server.
- Upon successfully connection, Soter will "Oper" up.
- When Soter is invited to a channel; it will immediately join.
- When Soter joins a channel for the first time it "Ops" itself.
- Soter then maintains a persistent "memory" of the channel's topic and modes.
- Soter will also maintain channel operators in the case of disruption.

## License

Soter is licensed under the MIT License.
