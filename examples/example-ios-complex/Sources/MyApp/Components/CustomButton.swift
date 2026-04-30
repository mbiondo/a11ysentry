import SwiftUI

struct CustomButton: View {
    var label: String
    var body: some View {
        Button(action: {}) {
            Text(label)
        }
        .accessibilityLabel("") // Shadowing the real label with empty string
    }
}
