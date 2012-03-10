#undef socklen_t
#undef u_int16_t
#undef u_int32_t
#undef u_int64_t
#undef u_int8_t

#undef in_addr_t

@BOTTOM@

/* Prototypes for missing functions */
#ifndef HAVE_STRLCAT
size_t	 strlcat(char *, const char *, size_t);
#endif

#ifndef HAVE_STRLCPY
size_t	 strlcpy(char *, const char *, size_t);
#endif
