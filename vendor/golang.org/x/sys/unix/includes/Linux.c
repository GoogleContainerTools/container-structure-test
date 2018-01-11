#define _LARGEFILE_SOURCE
#define _LARGEFILE64_SOURCE
#ifndef __LP64__
#define _FILE_OFFSET_BITS 64
#endif
#define _GNU_SOURCE

// <sys/ioctl.h> is broken on powerpc64, as it fails to include definitions of
// these structures. We just include them copied from <bits/termios.h>.
#if defined(__powerpc__)
struct sgttyb {
        char    sg_ispeed;
        char    sg_ospeed;
        char    sg_erase;
        char    sg_kill;
        short   sg_flags;
};

struct tchars {
        char    t_intrc;
        char    t_quitc;
        char    t_startc;
        char    t_stopc;
        char    t_eofc;
        char    t_brkc;
};

struct ltchars {
        char    t_suspc;
        char    t_dsuspc;
        char    t_rprntc;
        char    t_flushc;
        char    t_werasc;
        char    t_lnextc;
};
#endif

#include <bits/sockaddr.h>
#include <sys/epoll.h>
#include <sys/eventfd.h>
#include <sys/inotify.h>
#include <sys/ioctl.h>
#include <sys/mman.h>
#include <sys/mount.h>
#include <sys/prctl.h>
#include <sys/stat.h>
#include <sys/types.h>
#include <sys/time.h>
#include <sys/socket.h>
#include <sys/xattr.h>
#include <linux/if.h>
#include <linux/if_alg.h>
#include <linux/if_arp.h>
#include <linux/if_ether.h>
#include <linux/if_tun.h>
#include <linux/if_packet.h>
#include <linux/if_addr.h>
#include <linux/falloc.h>
#include <linux/filter.h>
#include <linux/fs.h>
#include <linux/keyctl.h>
#include <linux/netlink.h>
#include <linux/perf_event.h>
#include <linux/random.h>
#include <linux/reboot.h>
#include <linux/rtnetlink.h>
#include <linux/ptrace.h>
#include <linux/sched.h>
#include <linux/seccomp.h>
#include <linux/sockios.h>
#include <linux/wait.h>
#include <linux/icmpv6.h>
#include <linux/serial.h>
#include <linux/can.h>
#include <linux/vm_sockets.h>
#include <net/route.h>
#include <asm/termbits.h>

#ifndef MSG_FASTOPEN
#define MSG_FASTOPEN    0x20000000
#endif

#ifndef PTRACE_GETREGS
#define PTRACE_GETREGS	0xc
#endif

#ifndef PTRACE_SETREGS
#define PTRACE_SETREGS	0xd
#endif

#ifndef SOL_NETLINK
#define SOL_NETLINK	270
#endif

#ifdef SOL_BLUETOOTH
// SPARC includes this in /usr/include/sparc64-linux-gnu/bits/socket.h
// but it is already in bluetooth_linux.go
#undef SOL_BLUETOOTH
#endif

// Certain constants are missing from the fs/crypto UAPI
#define FS_KEY_DESC_PREFIX              "fscrypt:"
#define FS_KEY_DESC_PREFIX_SIZE         8
#define FS_MAX_KEY_SIZE                 64
