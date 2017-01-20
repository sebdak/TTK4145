
# Python 3.3.3 and 2.7.6
# python helloworld_python.py

from threading import Thread

count = 0

def someThreadFunction():
    global count
    print("Hello from a thread!")
    for i in xrange(1000000):
    	count+=1


def someThreadFunction2():
    global count
    print("Hello from a thread!")
    for i in xrange(1000000):
    	count-=1
# Potentially useful thing:
#   In Python you "import" a global variable, instead of "export"ing it when you declare it
#   (This is probably an effort to make you feel bad about typing the word "global")
    


def main():
    someThread = Thread(target = someThreadFunction, args = (),)
    someThread2 = Thread(target = someThreadFunction2, args = (),)

    someThread.start()
    someThread2.start()

    someThread.join()
    someThread2.join()
    print("Count equals: ", count)


main()