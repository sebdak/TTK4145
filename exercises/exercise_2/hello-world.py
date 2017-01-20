
# Python 3.3.3 and 2.7.6
# python helloworld_python.py

import threading

count = 0
lock = threading.Lock()

def someThreadFunction():
    global count
    lock.acquire()
    print("Hello from a thread!")
    for i in xrange(1000000):
    	count+=1
    lock.release()

def someThreadFunction2():
    global count
    lock.acquire()
    print("Hello from a thread!")
    for i in xrange(1000000):
        count-=1
    lock.release()
# Potentially useful thing:
#   In Python you "import" a global variable, instead of "export"ing it when you declare it
#   (This is probably an effort to make you feel bad about typing the word "global")
    


def main():
    someThread = threading.Thread(target = someThreadFunction, args = (),)
    someThread2 = threading.Thread(target = someThreadFunction2, args = (),)

    someThread.start()
    someThread2.start()

    someThread.join()
    someThread2.join()
    print("Count equals: ", count)


main()