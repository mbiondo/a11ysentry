import SwiftUI

struct ContentView: View {
    var body: some View {
        VStack {
            // Accessibility Issue: Image without accessibilityLabel
            Image("logo")
                .resizable()
                .frame(width: 100, height: 100)

            // Accessibility Issue: Button with no clear label
            Button(action: {}) {
                Image(systemName: "plus")
            }
            .accessibilityLabel("") // Empty label

            Text("Hello Semantix iOS!")
                .font(.title)
        }
        .padding()
    }
}
