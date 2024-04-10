#ifndef _HOOKS_NETWORK_RAW_H_
#define _HOOKS_NETWORK_RAW_H_

#include "helpers/network.h"
#include "perf_ring.h"

__attribute__((always_inline)) struct raw_packet_t *get_raw_packet_event() {
    u32 key = 0;
    return bpf_map_lookup_elem(&raw_packets, &key);
}

SEC("classifier/raw_packet")
int classifier_raw_packet(struct __sk_buff *skb) {
    struct packet_t *pkt = get_packet();
    if (pkt == NULL) {
        // should never happen
        return ACT_OK;
    }

    struct raw_packet_t *evt = get_raw_packet_event();
    if ((evt == NULL) || (skb == NULL)) {
        // should never happen
        return ACT_OK;
    }

    bpf_skb_pull_data(skb, 0);

    asm ("r1 = *(u32 *)(%[skb] + %[len_offset])\n\t"
         "if r1 <= 0 goto +1\n\t"
         "if r1 < %[limit] goto +2\n\t"
         "r0 = 0\n\t"
         "exit\n\t"
         :
         : [skb]"r"(skb), [len_offset]"i"(offsetof(struct __sk_buff, len)), [limit]"i"(sizeof(evt->data)));

#pragma clang diagnostic push
#pragma clang diagnostic ignored "-Wuninitialized"
    u32 len;
    asm ("r4 = r1");
    if (bpf_skb_load_bytes(skb, 0, evt->data, len) < 0) {
        return ACT_OK;
    }
#pragma clang diagnostic pop

    evt->len = skb->len;

    // process context
    fill_network_process_context(&evt->process, pkt);

    struct proc_cache_t *entry = get_proc_cache(evt->process.pid);
    if (entry == NULL) {
        evt->container.container_id[0] = 0;
    } else {
        copy_container_id_no_tracing(entry->container.container_id, &evt->container.container_id);
    }

    evt->flow = pkt->translated_ns_flow;

    unsigned int size2 = offsetof(struct raw_packet_t, data) + skb->len;
    if (size2 > sizeof(struct raw_packet_t)) {
        return ACT_OK;
    }
    send_event_with_size_ptr(skb, EVENT_RAW_PACKET, evt, size2);

    return ACT_OK;
}

#endif
