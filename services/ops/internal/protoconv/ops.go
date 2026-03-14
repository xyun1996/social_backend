package protoconv

import (
	opsv1 "github.com/xyun1996/social_backend/api/proto/ops/v1"
	opsservice "github.com/xyun1996/social_backend/services/ops/internal/service"
)

func ToProtoPresenceRecord(record opsservice.PresenceRecord) *opsv1.PresenceRecord {
	return &opsv1.PresenceRecord{
		PlayerId:        record.PlayerID,
		Status:          record.Status,
		SessionId:       record.SessionID,
		RealmId:         record.RealmID,
		Location:        record.Location,
		LastHeartbeatAt: record.LastHeartbeatAt,
		LastSeenAt:      record.LastSeenAt,
		ConnectedAt:     record.ConnectedAt,
		DisconnectedAt:  record.DisconnectedAt,
	}
}

func ToProtoPlayerOverview(overview opsservice.PlayerOverview) *opsv1.PlayerOverview {
	return &opsv1.PlayerOverview{
		PlayerId:           overview.PlayerID,
		Presence:           ToProtoPresenceRecord(overview.Presence),
		Friends:            append([]string(nil), overview.Friends...),
		Blocks:             append([]string(nil), overview.Blocks...),
		PendingInbox:       append([]string(nil), overview.PendingInbox...),
		PendingOutbox:      append([]string(nil), overview.PendingOutbox...),
		FriendCount:        int32(overview.FriendCnt),
		BlockCount:         int32(overview.BlockCnt),
		PendingInboxCount:  int32(overview.PendingInboxCount),
		PendingOutboxCount: int32(overview.PendingOutboxCount),
		CurrentPartyId:     overview.CurrentPartyID,
		CurrentGuildId:     overview.CurrentGuildID,
		CurrentGuildRole:   overview.CurrentGuildRole,
		CurrentQueueStatus: overview.CurrentQueueStatus,
	}
}

func ToProtoPartySnapshot(snapshot opsservice.PartySnapshot) *opsv1.PartySnapshot {
	members := make([]*opsv1.PartyMemberState, 0, len(snapshot.Members))
	for _, member := range snapshot.Members {
		members = append(members, &opsv1.PartyMemberState{
			PlayerId:  member.PlayerID,
			IsLeader:  member.IsLeader,
			IsReady:   member.IsReady,
			Presence:  member.Presence,
			SessionId: member.SessionID,
			RealmId:   member.RealmID,
			Location:  member.Location,
		})
	}
	record := &opsv1.PartySnapshot{
		PartyId: snapshot.PartyID,
		Count:   int32(snapshot.Count),
		Members: members,
	}
	if snapshot.Queue != nil {
		record.Queue = &opsv1.PartyQueueState{
			PartyId:   snapshot.Queue.PartyID,
			QueueName: snapshot.Queue.QueueName,
			Status:    snapshot.Queue.Status,
			JoinedBy:  snapshot.Queue.JoinedBy,
			JoinedAt:  snapshot.Queue.JoinedAt,
		}
	}
	return record
}

func ToProtoGuildSnapshot(snapshot opsservice.GuildSnapshot) *opsv1.GuildSnapshot {
	members := make([]*opsv1.GuildMemberState, 0, len(snapshot.Members))
	for _, member := range snapshot.Members {
		members = append(members, &opsv1.GuildMemberState{
			PlayerId:  member.PlayerID,
			Role:      member.Role,
			Presence:  member.Presence,
			SessionId: member.SessionID,
			RealmId:   member.RealmID,
			Location:  member.Location,
		})
	}
	logs := make([]*opsv1.GuildLogEntry, 0, len(snapshot.Logs))
	for _, entry := range snapshot.Logs {
		logs = append(logs, &opsv1.GuildLogEntry{
			Id:        entry.ID,
			Action:    entry.Action,
			ActorId:   entry.ActorID,
			TargetId:  entry.TargetID,
			Message:   entry.Message,
			CreatedAt: entry.CreatedAt,
		})
	}
	return &opsv1.GuildSnapshot{
		GuildId:               snapshot.GuildID,
		Name:                  snapshot.Name,
		OwnerId:               snapshot.OwnerID,
		Announcement:          snapshot.Announcement,
		AnnouncementUpdatedAt: snapshot.AnnouncementUpdatedAt,
		Count:                 int32(snapshot.Count),
		Members:               members,
		LogCount:              int32(snapshot.LogCount),
		Logs:                  logs,
	}
}

func ToProtoWorkerSnapshot(snapshot opsservice.WorkerSnapshot) *opsv1.WorkerSnapshot {
	jobs := make([]*opsv1.WorkerJob, 0, len(snapshot.Jobs))
	for _, job := range snapshot.Jobs {
		jobs = append(jobs, &opsv1.WorkerJob{
			Id:          job.ID,
			Type:        job.Type,
			Payload:     job.Payload,
			Status:      job.Status,
			Attempts:    int32(job.Attempts),
			LastError:   job.LastError,
			ClaimedBy:   job.ClaimedBy,
			CreatedAt:   job.CreatedAt,
			ClaimedAt:   job.ClaimedAt,
			CompletedAt: job.CompletedAt,
		})
	}
	return &opsv1.WorkerSnapshot{
		Status: snapshot.Status,
		Type:   snapshot.Type,
		Count:  int32(snapshot.Count),
		Jobs:   jobs,
	}
}

func ToProtoMySQLBootstrapSnapshot(snapshot opsservice.MySQLBootstrapSnapshot) *opsv1.MySQLBootstrapSnapshot {
	services := make([]*opsv1.MySQLBootstrapService, 0, len(snapshot.Services))
	for _, service := range snapshot.Services {
		services = append(services, &opsv1.MySQLBootstrapService{
			Service:      service.Service,
			Count:        int32(service.Count),
			MigrationIds: append([]string(nil), service.MigrationIDs...),
		})
	}
	return &opsv1.MySQLBootstrapSnapshot{
		Count:    int32(snapshot.Count),
		Services: services,
	}
}

func ToProtoRedisRuntimeSnapshot(snapshot opsservice.RedisRuntimeSnapshot) *opsv1.RedisRuntimeSnapshot {
	counters := make([]*opsv1.RedisWorkerStatusCount, 0, len(snapshot.WorkerStatusCounters))
	for _, counter := range snapshot.WorkerStatusCounters {
		counters = append(counters, &opsv1.RedisWorkerStatusCount{
			Status: counter.Status,
			Count:  int32(counter.Count),
		})
	}
	return &opsv1.RedisRuntimeSnapshot{
		RedisUrl:             snapshot.RedisURL,
		PresenceRecordCount:  int32(snapshot.PresenceRecordCount),
		GatewaySessionCount:  int32(snapshot.GatewaySessionCount),
		WorkerJobCount:       int32(snapshot.WorkerJobCount),
		WorkerStatusCounters: counters,
	}
}

func ToProtoDurableSummary(summary opsservice.DurableSummary) *opsv1.DurableSummary {
	out := &opsv1.DurableSummary{}
	if summary.MySQL != nil {
		out.Mysql = ToProtoMySQLBootstrapSnapshot(*summary.MySQL)
	}
	if summary.Redis != nil {
		out.Redis = ToProtoRedisRuntimeSnapshot(*summary.Redis)
	}
	return out
}
