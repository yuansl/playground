void test_empty_struct(void)
{
	struct {
		/* empty */
	} empty_structs[100];

	printf("sizeof(empty_structs)=%zd\n", sizeof empty_structs);
}

void test_unsigned_char(void)
{
	struct {
		short value;
	} x = { .value = 0xfefa };
	const char *p = (const char *)&x;

	for (size_t i = 0; i < sizeof(x); i++) {
		int8_t tmp = p[i];
		printf("p[%zd]= %s\n", i, tmp < 0 ? "true" : "false");
	}
}
