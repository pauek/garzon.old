
#include <vector>
#include <unistd.h>
using namespace std;

int main() { 
   vector<int> v(10000000);
   for (int i = 0; i < v.size(); i++) {
      v[i] = i;
   }
   v.clear();
   sleep(1);
   return 0; 
}
