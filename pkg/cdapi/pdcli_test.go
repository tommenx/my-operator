package cdapi

import "testing"

func TestParseData(t *testing.T) {
	str := `{
  "count": 8,
  "stores": [
    {
      "store": {
        "id": 1091,
        "address": "tidb-cluster-tikv-5.tidb-cluster-tikv-peer.default.svc:20160",
        "labels": [
          {
            "key": "host",
            "value": "worker137"
          }
        ],
        "version": "3.0.1",
        "state_name": "Up"
      },
      "status": {
        "capacity": "39 GiB",
        "available": "31 GiB",
        "leader_count": 29,
        "leader_weight": 1,
        "leader_score": 2595,
        "leader_size": 2595,
        "region_count": 88,
        "region_weight": 1,
        "region_score": 7931,
        "region_size": 7931,
        "start_ts": "2019-10-28T08:06:51Z",
        "last_heartbeat_ts": "2019-10-28T12:29:36.255271492Z",
        "uptime": "4h22m45.255271492s"
      }
    },
    {
      "store": {
        "id": 1092,
        "address": "tidb-cluster-tikv-4.tidb-cluster-tikv-peer.default.svc:20160",
        "labels": [
          {
            "key": "host",
            "value": "worker135"
          }
        ],
        "version": "3.0.1",
        "state_name": "Up"
      },
      "status": {
        "capacity": "39 GiB",
        "available": "31 GiB",
        "leader_count": 29,
        "leader_weight": 1,
        "leader_score": 2594,
        "leader_size": 2594,
        "region_count": 90,
        "region_weight": 1,
        "region_score": 7778,
        "region_size": 7778,
        "start_ts": "2019-10-28T08:06:59Z",
        "last_heartbeat_ts": "2019-10-28T12:29:32.750763745Z",
        "uptime": "4h22m33.750763745s"
      }
    },
    {
      "store": {
        "id": 1117,
        "address": "tidb-cluster-tikv-6.tidb-cluster-tikv-peer.default.svc:20160",
        "labels": [
          {
            "key": "host",
            "value": "worker135"
          }
        ],
        "version": "3.0.1",
        "state_name": "Up"
      },
      "status": {
        "capacity": "39 GiB",
        "available": "31 GiB",
        "leader_count": 32,
        "leader_weight": 1,
        "leader_score": 2589,
        "leader_size": 2589,
        "region_count": 97,
        "region_weight": 1,
        "region_score": 7794,
        "region_size": 7794,
        "start_ts": "2019-10-28T08:09:35Z",
        "last_heartbeat_ts": "2019-10-28T12:29:33.826605719Z",
        "uptime": "4h19m58.826605719s"
      }
    },
    {
      "store": {
        "id": 1122,
        "address": "tidb-cluster-tikv-7.tidb-cluster-tikv-peer.default.svc:20160",
        "labels": [
          {
            "key": "host",
            "value": "worker137"
          }
        ],
        "version": "3.0.1",
        "state_name": "Up"
      },
      "status": {
        "capacity": "39 GiB",
        "available": "31 GiB",
        "leader_count": 28,
        "leader_weight": 1,
        "leader_score": 2594,
        "leader_size": 2594,
        "region_count": 90,
        "region_weight": 1,
        "region_score": 7785,
        "region_size": 7785,
        "start_ts": "2019-10-28T08:09:42Z",
        "last_heartbeat_ts": "2019-10-28T12:29:32.279686075Z",
        "uptime": "4h19m50.279686075s"
      }
    },
    {
      "store": {
        "id": 1,
        "address": "tidb-cluster-tikv-0.tidb-cluster-tikv-peer.default.svc:20160",
        "labels": [
          {
            "key": "host",
            "value": "worker137"
          }
        ],
        "version": "3.0.1",
        "state_name": "Up"
      },
      "status": {
        "capacity": "39 GiB",
        "available": "29 GiB",
        "leader_count": 30,
        "leader_weight": 1,
        "leader_score": 2628,
        "leader_size": 2628,
        "region_count": 85,
        "region_weight": 1,
        "region_score": 7776,
        "region_size": 7776,
        "start_ts": "2019-10-28T07:17:28Z",
        "last_heartbeat_ts": "2019-10-28T12:29:34.406525631Z",
        "uptime": "5h12m6.406525631s"
      }
    },
    {
      "store": {
        "id": 4,
        "address": "tidb-cluster-tikv-2.tidb-cluster-tikv-peer.default.svc:20160",
        "labels": [
          {
            "key": "host",
            "value": "worker137"
          }
        ],
        "version": "3.0.1",
        "state_name": "Up"
      },
      "status": {
        "capacity": "39 GiB",
        "available": "29 GiB",
        "leader_count": 30,
        "leader_weight": 1,
        "leader_score": 2663,
        "leader_size": 2663,
        "region_count": 91,
        "region_weight": 1,
        "region_score": 7878,
        "region_size": 7878,
        "start_ts": "2019-10-28T07:17:29Z",
        "last_heartbeat_ts": "2019-10-28T12:29:32.66529485Z",
        "uptime": "5h12m3.66529485s"
      }
    },
    {
      "store": {
        "id": 6,
        "address": "tidb-cluster-tikv-3.tidb-cluster-tikv-peer.default.svc:20160",
        "labels": [
          {
            "key": "host",
            "value": "worker135"
          }
        ],
        "version": "3.0.1",
        "state_name": "Up"
      },
      "status": {
        "capacity": "39 GiB",
        "available": "29 GiB",
        "leader_count": 30,
        "leader_weight": 1,
        "leader_score": 2630,
        "leader_size": 2630,
        "region_count": 85,
        "region_weight": 1,
        "region_score": 7925,
        "region_size": 7925,
        "start_ts": "2019-10-28T07:17:42Z",
        "last_heartbeat_ts": "2019-10-28T12:29:37.007224696Z",
        "uptime": "5h11m55.007224696s"
      }
    },
    {
      "store": {
        "id": 7,
        "address": "tidb-cluster-tikv-1.tidb-cluster-tikv-peer.default.svc:20160",
        "labels": [
          {
            "key": "host",
            "value": "worker135"
          }
        ],
        "version": "3.0.1",
        "state_name": "Up"
      },
      "status": {
        "capacity": "39 GiB",
        "available": "27 GiB",
        "leader_count": 28,
        "leader_weight": 1,
        "leader_score": 2627,
        "leader_size": 2627,
        "region_count": 82,
        "region_weight": 1,
        "region_score": 7893,
        "region_size": 7893,
        "start_ts": "2019-10-28T07:17:43Z",
        "last_heartbeat_ts": "2019-10-28T12:29:36.99634924Z",
        "uptime": "5h11m53.99634924s"
      }
    }
  ]
}`
	res, err := ParseData([]byte(str))
	if err != nil {
		t.Errorf("parse error %+v", err)
	}
	t.Logf("res:%+v", res)
}

func TestGetRegionStatus(t *testing.T) {
	res, err := GetRegionStatus()
	if err != nil {
		t.Errorf("err=%+v", err)
	}
	t.Logf("res=%+v", res)
}
