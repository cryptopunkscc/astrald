package node

type LogConfig struct {
	IncludeEvents []string          `yaml:"include_events,omitempty"`
	ExcludeEvents []string          `yaml:"exclude_events,omitempty"`
	Level         int               `yaml:"level,omitempty"`
	TagLevels     map[string]int    `yaml:"tag_levels,omitempty"`
	TagColors     map[string]string `yaml:"tag_colors,omitempty"`
	HideDate      bool              `yaml:"hide_date,omitempty"`
}

func (c LogConfig) isIncluded(event string) bool {
	if len(c.IncludeEvents) == 0 {
		return false
	}
	for _, e := range c.IncludeEvents {
		if e == event {
			return true
		}
	}
	return false
}

func (c LogConfig) isExcluded(event string) bool {
	if len(c.ExcludeEvents) == 0 {
		return false
	}
	for _, e := range c.ExcludeEvents {
		if e == event {
			return true
		}
	}
	return false
}

func (c LogConfig) IsEventLoggable(event string) bool {
	if len(c.ExcludeEvents) > 0 {
		return !c.isExcluded(event)
	}

	return c.isIncluded(event)
}
