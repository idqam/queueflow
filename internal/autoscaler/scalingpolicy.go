package autoscaler

import "math"

type ScalingPolicy struct {
	JobsPerWorkerTarget int32
	MinReplicas         int32
	MaxReplicas         int32
}

func NewScalingPolicy(jobsPerWorkerTarget, minReplicas, maxReplicas int32) *ScalingPolicy {
	return &ScalingPolicy{
		JobsPerWorkerTarget: jobsPerWorkerTarget,
		MinReplicas:         minReplicas,
		MaxReplicas:         maxReplicas,
	}
}

func (p *ScalingPolicy) ComputeDesired(lag int64) int32 {
	if p == nil {
		return 1
	}

	jobsPerWorker := p.JobsPerWorkerTarget
	if jobsPerWorker <= 0 {
		jobsPerWorker = 20
	}

	if lag <= 0 {
		if p.MinReplicas > 0 {
			return p.MinReplicas
		}
		return 1
	}

	raw := int32(math.Ceil(float64(lag) / float64(jobsPerWorker)))
	if raw < p.MinReplicas {
		raw = p.MinReplicas
	}
	if p.MaxReplicas > 0 && raw > p.MaxReplicas {
		raw = p.MaxReplicas
	}

	return raw
}
