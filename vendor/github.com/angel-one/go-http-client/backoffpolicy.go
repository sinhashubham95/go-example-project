package httpclient

import "time"

// BackoffPolicy is the type for backoff policy
type BackoffPolicy struct {
	constantBackoff    *ConstantBackoff
	exponentialBackoff *ExponentialBackoff
}

// NewBackoffPolicy is used to create a new backoff policy
func NewBackoffPolicy(configMap map[string]interface{}) *BackoffPolicy {
	backoffPolicy := &BackoffPolicy{}

	constantBackoffMap, err := getConfigOptionMap(configMap, "constantbackoff")
	if err == nil {
		backoffPolicy.constantBackoff = NewConstantBackoff(constantBackoffMap)
	}

	exponentialBackoffMap, err := getConfigOptionMap(configMap, "exponentialbackoff")
	if err == nil {
		backoffPolicy.exponentialBackoff = NewExponentialBackoff(exponentialBackoffMap)
	}

	return backoffPolicy
}

// SetConstantBackoff is used to use the constant backoff policy
func (bop *BackoffPolicy) SetConstantBackoff(constantBackoff *ConstantBackoff) *BackoffPolicy {
	bop.constantBackoff = constantBackoff
	return bop
}

// SetExponentialBackoff is used to set the exponential backoff policy
func (bop *BackoffPolicy) SetExponentialBackoff(exponentialBackoff *ExponentialBackoff) *BackoffPolicy {
	bop.exponentialBackoff = exponentialBackoff
	return bop
}

// ConstantBackoff is used to create a new constant backoff
type ConstantBackoff struct {
	interval              time.Duration
	maximumJitterInterval time.Duration
}

// NewConstantBackoff is used to create a new constant backoff
func NewConstantBackoff(configMap map[string]interface{}) *ConstantBackoff {
	constantBackoff := &ConstantBackoff{}

	constantBackoffInterval, err := getConfigOptionInt(configMap, "intervalinmillis")
	if err == nil {
		constantBackoff.interval = time.Duration(constantBackoffInterval) * time.Millisecond
	}
	maxJitterInterval, err := getConfigOptionInt(configMap, "maxjitterintervalinmillis")
	if err == nil {
		constantBackoff.maximumJitterInterval = time.Duration(maxJitterInterval) * time.Millisecond
	}

	return constantBackoff
}

// SetInterval is used to set the interval for the backoff
func (cb *ConstantBackoff) SetInterval(interval time.Duration) *ConstantBackoff {
	cb.interval = interval
	return cb
}

// SetMaximumJitterInterval is used to set the jitter interval
func (cb *ConstantBackoff) SetMaximumJitterInterval(maximumJitterInterval time.Duration) *ConstantBackoff {
	cb.maximumJitterInterval = maximumJitterInterval
	return cb
}

// ExponentialBackoff is used to create a new exponential backoff
type ExponentialBackoff struct {
	exponentFactor        float64
	initialTimeout        time.Duration
	maxTimeout            time.Duration
	maximumJitterInterval time.Duration
}

// NewExponentialBackoff is used to create a new exponential backoff
func NewExponentialBackoff(configMap map[string]interface{}) *ExponentialBackoff {
	exponentialBackoff := &ExponentialBackoff{}
	exponentialBackoff.exponentFactor, _ = getConfigOptionFloat(configMap, "exponentfactor")
	initialTimeout, err := getConfigOptionInt(configMap, "initialtimeoutinmillis")
	if err == nil {
		exponentialBackoff.initialTimeout = time.Duration(initialTimeout) * time.Millisecond
	}
	maxTimeout, err := getConfigOptionInt(configMap, "maxtimeoutinmillis")
	if err == nil {
		exponentialBackoff.maxTimeout = time.Duration(maxTimeout) * time.Millisecond
	}
	maximumJitterInterval, err := getConfigOptionInt(configMap, "maxjitterintervalinmillis")
	if err == nil {
		exponentialBackoff.maximumJitterInterval = time.Duration(maximumJitterInterval) * time.Millisecond
	}
	return exponentialBackoff
}

// SetExponentFactor is used to set the exponential factor for backoff
func (eb *ExponentialBackoff) SetExponentFactor(exponentFactor float64) *ExponentialBackoff {
	eb.exponentFactor = exponentFactor
	return eb
}

// SetMaxTimeout is used to set the maximum timeout duration
func (eb *ExponentialBackoff) SetMaxTimeout(maxTimeout time.Duration) *ExponentialBackoff {
	eb.maxTimeout = maxTimeout
	return eb
}

// SetMaximumJitterInterval is used to set the jitter interval
func (eb *ExponentialBackoff) SetMaximumJitterInterval(maximumJitterInterval time.Duration) *ExponentialBackoff {
	eb.maximumJitterInterval = maximumJitterInterval
	return eb
}

// SetInitialTimeout is used to set the initial wait time before the backoff
func (eb *ExponentialBackoff) SetInitialTimeout(initialTimeout time.Duration) *ExponentialBackoff {
	eb.initialTimeout = initialTimeout
	return eb
}
