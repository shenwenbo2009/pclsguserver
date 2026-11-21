package metrics

import (
	"sync"
	"time"

	"pclsguserver/internal/model"
	"pclsguserver/internal/repo"

	"github.com/google/uuid"
)

type Service struct {
	repo    *repo.MetricsRepo
	eventCh chan *model.MetricsEventRaw
	batch   []*model.MetricsEventRaw
	mu      sync.Mutex
}

func NewService(metricsRepo *repo.MetricsRepo) *Service {
	s := &Service{
		repo:    metricsRepo,
		eventCh: make(chan *model.MetricsEventRaw, 1024),
	}
	go s.processEvents()
	return s
}

func (s *Service) RecordEvent(dimension model.MetricsDimension, subjectKey, clientId string) {
	ev := &model.MetricsEventRaw{
		ID:         uuid.NewString(),
		Dimension:  dimension,
		SubjectKey: subjectKey,
		ClientID:   clientId,
		OccurredAt: time.Now(),
	}
	select {
	case s.eventCh <- ev:
	default:
	}
}

func (s *Service) processEvents() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case ev, ok := <-s.eventCh:
			if !ok {
				return
			}
			s.mu.Lock()
			s.batch = append(s.batch, ev)
			if len(s.batch) >= 100 {
				s.flush()
			}
			s.mu.Unlock()
		case <-ticker.C:
			s.mu.Lock()
			s.flush()
			s.mu.Unlock()
		}
	}
}

func (s *Service) flush() {
	if len(s.batch) == 0 {
		return
	}
	for _, ev := range s.batch {
		_ = s.repo.AppendRaw(nil, ev)
	}
	s.batch = s.batch[:0]
}

func (s *Service) Aggregate() error {
	events, err := s.repo.ListUnaggregated(10000)
	if err != nil {
		return err
	}
	if len(events) == 0 {
		return nil
	}

	type aggKey struct {
		dimension   string
		subjectKey  string
		granularity string
		bucket      time.Time
	}
	aggMap := make(map[aggKey]*model.MetricsEvent)
	clientSets := make(map[aggKey]map[string]bool)

	for _, ev := range events {
		for _, gran := range []model.Granularity{model.GranularityHour, model.GranularityDay} {
			bucket := bucketOf(ev.OccurredAt, gran)
			key := aggKey{
				dimension:   string(ev.Dimension),
				subjectKey:  ev.SubjectKey,
				granularity: string(gran),
				bucket:      bucket,
			}
			if _, ok := aggMap[key]; !ok {
				aggMap[key] = &model.MetricsEvent{
					ID:          uuid.NewString(),
					Dimension:   ev.Dimension,
					SubjectKey:  ev.SubjectKey,
					Granularity: gran,
					BucketStart: bucket,
					BucketEnd:   bucketEnd(bucket, gran),
				}
				clientSets[key] = make(map[string]bool)
			}
			aggMap[key].Count++
			if ev.ClientID != "" {
				if !clientSets[key][ev.ClientID] {
					clientSets[key][ev.ClientID] = true
					aggMap[key].UniqueClientCount++
				}
			}
		}
	}

	for _, e := range aggMap {
		_ = s.repo.UpsertAggregated(nil, e)
	}

	ids := make([]string, len(events))
	for i, ev := range events {
		ids[i] = ev.ID
	}
	return s.repo.MarkAggregated(ids)
}

func bucketOf(t time.Time, gran model.Granularity) time.Time {
	switch gran {
	case model.GranularityHour:
		return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), 0, 0, 0, t.Location())
	case model.GranularityDay:
		return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	}
	return t
}

func bucketEnd(start time.Time, gran model.Granularity) time.Time {
	switch gran {
	case model.GranularityHour:
		return start.Add(time.Hour)
	case model.GranularityDay:
		return start.Add(24 * time.Hour)
	}
	return start
}

func (s *Service) Query(dimension string, from, to time.Time, granularity, subjectKey string) ([]model.MetricsEvent, error) {
	return s.repo.QueryAggregated(dimension, subjectKey, from, to, granularity)
}

func (s *Service) Stop() {
	close(s.eventCh)
}
