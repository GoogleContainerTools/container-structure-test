#include <sys/capability.h>
#include <sys/param.h>
#include <sys/types.h>
#include <sys/event.h>
#include <sys/socket.h>
#include <sys/sockio.h>
#include <sys/sysctl.h>
#include <sys/mman.h>
#include <sys/wait.h>
#include <sys/ioctl.h>
#include <net/bpf.h>
#include <net/if.h>
#include <net/if_types.h>
#include <net/route.h>
#include <netinet/in.h>
#include <termios.h>
#include <netinet/ip.h>
#include <netinet/ip_mroute.h>
#include <sys/extattr.h>

#if __FreeBSD__ >= 10
#define IFT_CARP	0xf8	// IFT_CARP is deprecated in FreeBSD 10
#undef SIOCAIFADDR
#define SIOCAIFADDR	_IOW(105, 26, struct oifaliasreq)	// ifaliasreq contains if_data
#undef SIOCSIFPHYADDR
#define SIOCSIFPHYADDR	_IOW(105, 70, struct oifaliasreq)	// ifaliasreq contains if_data
#endif
