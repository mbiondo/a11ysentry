import SwiftUI

struct ContentView: View {
    var body: some View {
        ScrollView {
            VStack(spacing: 20) {
                Text("Complex iOS A11y Test")
                    .font(.largeTitle)
                
                // Rule: WCAG 1.1.1 (Image without label) - Error
                Image("logo")
                    .resizable()
                    .frame(width: 50, height: 50)
                
                // Rule: WCAG 1.1.1 (Image with explicit label) - OK
                Image("banner")
                    .resizable()
                    .frame(height: 100)
                    .accessibilityLabel("Welcome Banner")
                
                // Rule: WCAG 4.1.2 (Button with implicit label from Text) - OK
                Button(action: {}) {
                    Text("Login Now")
                }
                
                // Rule: WCAG 4.1.2 (Button with only an icon and no label) - Error
                Button(action: {}) {
                    Image(systemName: "plus")
                }
                
                // Rule: accessibilityElement(children: .combine)
                // This combines the labels of all children into one.
                VStack {
                    Text("Item Name")
                    Text("Item Description")
                }
                .accessibilityElement(children: .combine)
                
                // Rule: Hidden Focus equivalent (accessibilityHidden)
                Button(action: {}) {
                    Text("I am hidden from VoiceOver")
                }
                .accessibilityHidden(true)
            }
            .padding()
        }
    }
}

struct ContentView_Previews: PreviewProvider {
    static var previews: some View {
        ContentView()
    }
}
