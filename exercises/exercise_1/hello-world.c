// gcc 4.7.2 +
// gcc -std=gnu99 -Wall -g -o helloworld_c helloworld_c.c -lpthread

#include <pthread.h>
#include <stdio.h>

int count = 0;
// Note the return type: void*
void* someThreadFunction(){
	int i;
    for(i = 0; i < 1000000; ++i){
    	count++;
    }
    return NULL;
}

void* someThreadFunction2(){
    int i;
    for(i = 0; i < 1000000; ++i){
    	count--;
    }
    return NULL;
}


int main(){
    pthread_t someThread;
    pthread_create(&someThread, NULL, someThreadFunction, NULL);
    // Arguments to a thread would be passed here ---------^
    pthread_t someThread2;
    pthread_create(&someThread2, NULL, someThreadFunction2, NULL);
    
    pthread_join(someThread, NULL);
    pthread_join(someThread2, NULL);

    printf("count equals: %d",count);
    return 0;
    
}