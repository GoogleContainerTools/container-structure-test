#include <sys/types.h>
#include <sys/param.h>
#include <sys/event.h>
#include <sys/mman.h>
#include <sys/socket.h>
#include <sys/sockio.h>
#include <sys/sysctl.h>
#include <sys/termios.h>
#include <sys/ttycom.h>
#include <sys/wait.h>
#include <net/bpf.h>
#include <net/if.h>
#include <net/if_types.h>
#include <net/if_var.h>
#include <net/route.h>
#include <netinet/in.h>
#include <netinet/in_systm.h>
#include <netinet/ip.h>
#include <netinet/ip_mroute.h>
#include <netinet/if_ether.h>
#include <net/if_bridge.h>

// We keep some constants not supported in OpenBSD 5.5 and beyond for
// the promise of compatibility.
#define EMUL_ENABLED		0x1
#define EMUL_NATIVE		0x2
#define IPV6_FAITH		0x1d
#define IPV6_OPTIONS		0x1
#define IPV6_RTHDR_STRICT	0x1
#define IPV6_SOCKOPT_RESERVED1	0x3
#define SIOCGIFGENERIC		0xc020693a
#define SIOCSIFGENERIC		0x80206939
#define WALTSIG			0x4
