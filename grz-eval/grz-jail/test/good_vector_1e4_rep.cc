
#include <vector>
using namespace std;

int main() { 
   vector<int> v;
   for (int i = 0; i < 1000; i++) {
      v = vector<int>(10000);
   }
   return 0;
}
