// +build ignore

#include "common.h"
#include "bpf_helpers.h"

char __license[] SEC("license") = "Dual MIT/GPL";

// struct bpf_map_def SEC("maps") pkt_count = {
//     .type = BPF_MAP_TYPE_ARRAY,
//     .key_size = sizeof(u32),
//     .value_size = sizeof(u64),
//     .max_entries = 1,
// };


struct {
	__uint(type, BPF_MAP_TYPE_RINGBUF);
	__uint(max_entries, 1 << 24);
} events SEC(".maps");

struct event_t {
	__u16 sport;
	__be16 dport;
	__be32 saddr;
	__be32 daddr;
};

#define TASK_CUSTOM_LEN 32

struct eventPayload_t {
    char custom[TASK_CUSTOM_LEN];
};

struct __sk_buff {
	__u32 len;
	__u32 pkt_type;
	__u32 mark;
	__u32 queue_mapping;
	__u32 protocol;
	__u32 vlan_present;
	__u32 vlan_tci;
	__u32 vlan_proto;
	__u32 priority;
	__u32 ingress_ifindex;
	__u32 ifindex;
	__u32 tc_index;
	__u32 cb[5];
	__u32 hash;
	__u32 tc_classid;
	__u32 data;
	__u32 data_end;
	__u32 napi_id;

	/* Accessed by BPF_PROG_TYPE_sk_skb types from here to ... */
	__u32 family;
	__u32 remote_ip4;	/* Stored in network byte order */
	__u32 local_ip4;	/* Stored in network byte order */
	__u32 remote_ip6[4];	/* Stored in network byte order */
	__u32 local_ip6[4];	/* Stored in network byte order */
	__u32 remote_port;	/* Stored in network byte order */
	__u32 local_port;	/* stored in host byte order */
	/* ... here. */

	__u32 data_meta;
//	__bpf_md_ptr(struct bpf_flow_keys *, flow_keys);
	__u64 tstamp;
	__u32 wire_len;
	__u32 gso_segs;
//	__bpf_md_ptr(struct bpf_sock *, sk);
};

// static inline __attribute__((always_inline))
// u32 bpf_ntohl(u32 val) {
//   /* gcc will use bswapsi2 insn */
//   return __builtin_bswap32(val);
// }

static inline __attribute__((always_inline))
u16 bpf_ntohs(u16 val) {
  /* will be recognized by gcc into rotate insn and eventually rolw 8 */
  return (val << 8) | (val >> 8);
}

#define BPF_PACKET_HEADER __attribute__((packed))

struct ethernet_t {
  unsigned long long  dst:48;
  unsigned long long  src:48;
  unsigned int        type:16;
} BPF_PACKET_HEADER;

struct ip_t {
  unsigned char   ver:4;           // byte 0
  unsigned char   hlen:4;
  unsigned char   tos;
  unsigned short  tlen;
  unsigned short  identification; // byte 4
  unsigned short  ffo_unused:1;
  unsigned short  df:1;
  unsigned short  mf:1;
  unsigned short  foffset:13;
  unsigned char   ttl;             // byte 8
  unsigned char   nextp;
  unsigned short  hchecksum;
  unsigned int    src;            // byte 12
  unsigned int    dst;            // byte 16
} BPF_PACKET_HEADER;

struct tcp_t {
  unsigned short  src_port;   // byte 0
  unsigned short  dst_port;
  unsigned int    seq_num;    // byte 4
  unsigned int    ack_num;    // byte 8
  unsigned char   offset:4;    // byte 12
  unsigned char   reserved:4;
  unsigned char   flag_cwr:1;
  unsigned char   flag_ece:1;
  unsigned char   flag_urg:1;
  unsigned char   flag_ack:1;
  unsigned char   flag_psh:1;
  unsigned char   flag_rst:1;
  unsigned char   flag_syn:1;
  unsigned char   flag_fin:1;
  unsigned short  rcv_wnd;
  unsigned short  cksum;      // byte 16
  unsigned short  urg_ptr;
} BPF_PACKET_HEADER;

#define ETH_HLEN 14
#define IP_HLEN 20

#define IP_TCP 	6

// packet parsing state machine helpers
#define cursor_advance(_cursor, _len) \
  ({ void *_tmp = _cursor; _cursor += _len; _tmp; })

/* Packet types.  */

#define PACKET_HOST		0		/* To us.  */
#define PACKET_BROADCAST	1		/* To all.  */
#define PACKET_MULTICAST	2		/* To group.  */
#define PACKET_OTHERHOST	3		/* To someone else.  */
#define PACKET_OUTGOING		4		/* Originated by us . */
#define PACKET_LOOPBACK		5
#define PACKET_FASTROUTE	6

void ringbuf_submit_payload(char *payload, u8 len) {
	struct eventPayload_t *ep;
	ep = bpf_ringbuf_reserve(&events, sizeof(struct eventPayload_t), 0);
	if (!ep) {
		return;
	}

    for (int i = 0; i < len; i++) {
        ep->custom[i] = payload[i];
    }

	bpf_ringbuf_submit(ep, 0);
}

void ringbuf_submit_header(struct __sk_buff *skb) {
	struct event_t *tcp_info;
	tcp_info = bpf_ringbuf_reserve(&events, sizeof(struct event_t), 0);
	if (!tcp_info) {
		return;
	}

    tcp_info->saddr = skb->local_ip4;
    tcp_info->sport = skb->local_port;

    tcp_info->daddr = skb->remote_ip4;
    tcp_info->dport = bpf_ntohs(skb->remote_port >> 16);

    if (skb->pkt_type == PACKET_HOST) {
        // need to swap src with dst
        __be32 tmp32 = tcp_info->saddr;
        tcp_info->saddr = tcp_info->daddr;
        tcp_info->daddr = tmp32;
        __u16 tmp16 = tcp_info->sport;
        tcp_info->sport = tcp_info->dport;
        tcp_info->dport = tmp16;
    }

	bpf_ringbuf_submit(tcp_info, 0);
}

#define ETH_P_IP	0x0800		/* Internet Protocol packet	*/
#define ETH_P_IPV6	0x86DD		/* IPv6 over bluebook		*/

static inline __attribute__((always_inline)) int dump_skb_packet(struct __sk_buff* skb) {
    if ( bpf_ntohs(skb->remote_port >> 16) != 80 ) {
        return 1;
    }

    u16 offset = IP_HLEN; // start at the next header (should be TCP header)
    u8 tcp_data_offset = 0;
    bpf_skb_load_bytes(skb, offset+12, (void*)&tcp_data_offset, 1);
    offset += 4*(tcp_data_offset>>4 & 0x0f); // now we're at the actual data

    //load first 7 byte of payload into p (payload_array)
    //direct access to skb not allowed
    char p[7];
    bpf_skb_load_bytes(skb, offset, (void*)&p[0], 7);

    u8 haveHTTP = 0;

    //find a match with an HTTP message
    //HTTP
    if ((p[0] == 'H') && (p[1] == 'T') && (p[2] == 'T') && (p[3] == 'P')) {
        haveHTTP = 1;
    }
    //GET
    if ((p[0] == 'G') && (p[1] == 'E') && (p[2] == 'T')) {
        haveHTTP = 1;
    }
    //POST
    if ((p[0] == 'P') && (p[1] == 'O') && (p[2] == 'S') && (p[3] == 'T')) {
        haveHTTP = 1;
    }
    //HEAD
    if ((p[0] == 'H') && (p[1] == 'E') && (p[2] == 'A') && (p[3] == 'D')) {
        haveHTTP = 1;
    }

    if (haveHTTP != 1) {
        return 1;
    }

    ringbuf_submit_header(skb);

    char fullbuff[TASK_CUSTOM_LEN+4]; // data buffer with 4-byte prefix
    char *pbuff = &fullbuff[4];
    u8 finished = 0;

    fullbuff[0] = 0;
    fullbuff[1] = 0;
    fullbuff[2] = 0;
    fullbuff[3] = 0;

    for (int cnt = 0; cnt < 10 && !finished; cnt += 1) {
        u8 len = TASK_CUSTOM_LEN;
        bpf_skb_load_bytes(skb, offset + cnt * TASK_CUSTOM_LEN, &pbuff[0], TASK_CUSTOM_LEN);

        // see if we're reached the end of HTTP headers block (\r\n\r\n)
        for (int idx = 0; idx < TASK_CUSTOM_LEN; idx++) {
            if (fullbuff[idx] == '\r' && fullbuff[idx+1] == '\n' && fullbuff[idx+2] == '\r' && fullbuff[idx+3] == '\n') {
                finished = 1;
                fullbuff[idx+2] = 0;
                len = idx;
                break;
            }
        }

        if (len) {
            ringbuf_submit_payload(pbuff, len);
        }

        // store the last 4 bytes in our buffer prefix space
        fullbuff[0] = pbuff[TASK_CUSTOM_LEN-4];
        fullbuff[1] = pbuff[TASK_CUSTOM_LEN-3];
        fullbuff[2] = pbuff[TASK_CUSTOM_LEN-2];
        fullbuff[3] = pbuff[TASK_CUSTOM_LEN-1];
    }
    return 1;
}

SEC("cgroup_skb/egress")
int dump_egress_packets(struct __sk_buff *skb) {
    return dump_skb_packet(skb);
}

SEC("cgroup_skb/ingress")
int dump_ingress_packets(struct __sk_buff *skb) {
    return dump_skb_packet(skb);
}
