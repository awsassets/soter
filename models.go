package main

import (
	"time"
)

type Channel struct {
	ID        int       `storm:"id,increment"`
	Name      string    `storm:"index"`
	CreatedAt time.Time `storm:"index"`
}

type Mode struct {
	ID        int    `storm:"id,increment"`
	Channel   string `storm:"index"`
	Modes     []string
	CreatedAt time.Time `storm:"index"`
	UpdatedAt time.Time `storm:"index"`
}

type Topic struct {
	ID        int    `storm:"id,increment"`
	Channel   string `storm:"index"`
	Topic     string
	CreatedAt time.Time `storm:"index"`
	UpdatedAt time.Time `storm:"index"`
}

func NewChannel(name string) Channel {
	return Channel{
		Name:      name,
		CreatedAt: time.Now(),
	}
}

func NewMode(channel string, modes []string) Mode {
	return Mode{
		Channel:   channel,
		Modes:     modes,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func (m *Mode) Equal(modes []string) bool {
	if m.Modes == nil && modes == nil {
		return true
	}

	if m.Modes == nil || modes == nil {
		return false
	}

	if len(m.Modes) != len(modes) {
		return false
	}

	for i := range m.Modes {
		if m.Modes[i] != modes[i] {
			return false
		}
	}

	return true
}

func (m *Mode) AppendModes(modes []string) {
	m.Modes = append(m.Modes, modes...)
	m.UpdatedAt = time.Now()
}

func NewTopic(channel, topic string) Topic {
	return Topic{
		Channel:   channel,
		Topic:     topic,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func (t *Topic) SetTopic(topic string) {
	t.Topic = topic
	t.UpdatedAt = time.Now()
}
