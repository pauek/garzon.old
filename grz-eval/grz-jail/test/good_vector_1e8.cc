
#include <vector>
using namespace std;

int main() { 
   vector<int> v(100000000);
   for (int i = 0; i < v.size(); i++) {
      v[i] = i;
   }
   return 0; 
}
