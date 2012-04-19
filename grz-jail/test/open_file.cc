
#include <fstream>
using namespace std;

int main() {
   ofstream fout("/tmp/test");
   fout << "Hi, there!" << endl;
}
