binaries=makeSet1

# Needed for Linux to get rid of 'Value too large for defined data type' error.
CPPFLAGS += -D_FILE_OFFSET_BITS=64
# Not always needed
CPPFLAGS += -lrt

all: $(binaries)

clean:
	rm $(binaries)

