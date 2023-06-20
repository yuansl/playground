extern int foo;

int bar(int);

int foof(void)
{
	return bar(foo);
}
