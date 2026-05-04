import SwiftUI
import MyApp.Components // Mocking module import to test resolution

struct MainView: View {
    var body: some View {
        VStack {
            EnvironmentConfig()
            CustomButton(label: "Click Me")
        }
    }
}
