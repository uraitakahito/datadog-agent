#ifndef __KAFKA_PKTBUF_H
#define __KAFKA_PKTBUF_H

#include "protocols/helpers/big_endian.h"
#include "protocols/kafka/defs.h"
#include "protocols/kafka/maps.h"
#include "protocols/kafka/types.h"
#include "protocols/read_into_buffer.h"

enum pktbuf_type {
    PKTBUF_SKB,
    PKTBUF_TLS,
    PKTBUF_SK_MSG,
};

struct pktbuf {
    enum pktbuf_type type;
    union {
        struct {
            struct __sk_buff *skb;
            skb_info_t *skb_info;
        };
        struct {
            tls_dispatcher_arguments_t *tls;
        };
        struct {
            struct sk_msg_md *sk_msg_md;
        };
    };
};

typedef const struct pktbuf pktbuf_t;

// Never defined, intended to catch some implementation/usage errors at build-time.
extern void pktbuf_invalid_operation(void);

static __always_inline __maybe_unused u32 pktbuf_data_offset(pktbuf_t pkt)
{
    switch (pkt.type) {
    case PKTBUF_SKB:
        return pkt.skb_info ? pkt.skb_info->data_off : 0;
    case PKTBUF_TLS:
        return pkt.tls->data_off;
    case PKTBUF_SK_MSG:
        return 0;
    }

    pktbuf_invalid_operation();
    return 0;
}

static __always_inline __maybe_unused u32 pktbuf_data_end(pktbuf_t pkt)
{
    switch (pkt.type) {
    case PKTBUF_SKB:
        return pkt.skb_info ? pkt.skb_info->data_end : 0;
    case PKTBUF_TLS:
        return pkt.tls->data_end;
    case PKTBUF_SK_MSG:
        return pkt.sk_msg_md->size;
    }

    pktbuf_invalid_operation();
    return 0;
}

static __always_inline long pktbuf_load_bytes_with_telemetry(pktbuf_t pkt, u32 offset, void *to, u32 len)
{
    switch (pkt.type) {
    case PKTBUF_SKB:
        return bpf_skb_load_bytes_with_telemetry(pkt.skb, offset, to, len);
    case PKTBUF_TLS:
        return bpf_probe_read_user(to, len, pkt.tls->buffer_ptr + offset);
    case PKTBUF_SK_MSG:
        return bpf_sk_msg_load_bytes(pkt.sk_msg_md, offset, to, len);
    }

    pktbuf_invalid_operation();
    return 0;
}

static __always_inline __maybe_unused long pktbuf_load_bytes(pktbuf_t pkt, u32 offset, void *to, u32 len)
{
    switch (pkt.type) {
    case PKTBUF_SKB:
        return bpf_skb_load_bytes(pkt.skb, offset, to, len);
    case PKTBUF_TLS:
        return bpf_probe_read_user(to, len, pkt.tls->buffer_ptr + offset);
    case PKTBUF_SK_MSG:
        return pktbuf_load_bytes_with_telemetry(pkt, offset, to, len);
    }

    pktbuf_invalid_operation();
    return 0;
}

static __always_inline pktbuf_t pktbuf_from_skb(struct __sk_buff* skb, skb_info_t *skb_info)
{
    return (pktbuf_t) {
        .type = PKTBUF_SKB,
        .skb = skb,
        .skb_info = skb_info,
    };
}

static __always_inline __maybe_unused pktbuf_t pktbuf_from_tls(tls_dispatcher_arguments_t *tls)
{
    return (pktbuf_t) {
        .type = PKTBUF_TLS,
        .tls = tls,
    };
}

static __always_inline __maybe_unused pktbuf_t pktbuf_from_sk_msg_md(struct sk_msg_md *msg)
{
    return (pktbuf_t) {
        .type = PKTBUF_SK_MSG,
        .sk_msg_md = msg,
    };
}

#define PKTBUF_READ_BIG_ENDIAN(type_)                                                                      \
    static __always_inline __maybe_unused bool pktbuf_read_big_endian_##type_(pktbuf_t pkt, u32 offset, type_ *out) {  \
        switch (pkt.type) { \
        case PKTBUF_SKB: \
            return read_big_endian_##type_(pkt.skb, offset, out); \
        case PKTBUF_TLS: \
            return read_big_endian_user_##type_(pkt.tls->buffer_ptr, pkt.tls->data_end, offset, out); \
        case PKTBUF_SK_MSG: \
            return read_big_endian_sk_msg_##type_(pkt.sk_msg_md, offset, out); \
        } \
        pktbuf_invalid_operation(); \
        return false; \
    }

PKTBUF_READ_BIG_ENDIAN(s32)
PKTBUF_READ_BIG_ENDIAN(s16)
PKTBUF_READ_BIG_ENDIAN(s8)

#define PKTBUF_READ_INTO_BUFFER(name, total_size, blk_size)     \
    READ_INTO_USER_BUFFER(name, total_size)                    \
    READ_INTO_BUFFER(name, total_size, blk_size)               \
    READ_INTO_BUFFER_SK_MSG(name, total_size, blk_size)        \
    static __always_inline void pktbuf_read_into_buffer_##name(char *buffer, pktbuf_t pkt, u32 offset) { \
        switch (pkt.type) { \
        case PKTBUF_SKB:        \
            read_into_buffer_##name(buffer, pkt.skb, offset);                              \
            return; \
        case PKTBUF_TLS: \
            read_into_user_buffer_##name(buffer, pkt.tls->buffer_ptr + pkt.tls->data_off + offset); \
            return; \
        case PKTBUF_SK_MSG:        \
            read_into_buffer_sk_msg_##name(buffer, pkt.sk_msg_md, offset);                              \
            return; \
        } \
        pktbuf_invalid_operation(); \
    }

// Wraps the mechanism of reading big-endian number (s16 or s32) from the packet, and increasing the offset.
#define PKTBUF_READ_BIG_ENDIAN_WRAPPER(type, name, pkt, offset)        \
    type name = 0;                                              \
    if (!pktbuf_read_big_endian_##type(pkt, offset, &name)) {   \
        return false;                                           \
    }                                                           \
    offset += sizeof(type);

#endif