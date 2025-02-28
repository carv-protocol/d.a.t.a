package core

import (
	"fmt"
	"strings"
	"time"
)

type Limitation struct {
	Enable            bool
	MinTokenBalance   float64
	FrequencyFactor   float64
	FrequencyDuration int64
}

func (l Limitation) CheckLimit(stakeholder *Stakeholder) (string, bool) {
	if !l.Enable {
		return "", false
	}

	if stakeholder.TokenBalance.Balance < l.MinTokenBalance {
		return fmt.Sprintf("The balance of $CARV must be greater than %f", l.MinTokenBalance), true
	}

	var (
		now      = time.Now()
		duration = time.Second * time.Duration(l.FrequencyDuration)
		reqCount int
	)
	for _, historicalMsg := range stakeholder.HistoricalMsgs {
		if strings.Contains(historicalMsg.Content, stakeholder.ID) &&
			now.Sub(time.Unix(historicalMsg.Timestamp, 0)) < duration {
			reqCount++
		}
	}

	maxReqCount := int(stakeholder.TokenBalance.Balance / l.FrequencyFactor)
	if reqCount >= maxReqCount {
		return fmt.Sprintf("Talk frequency is too high, you can talk to AI up to %d times in %s",
			maxReqCount, duration.String()), true
	}

	return "", false
}
