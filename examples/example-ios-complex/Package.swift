// swift-tools-version: 5.9
import PackageDescription

let package = Package(
    name: "MyAppComplex",
    platforms: [.iOS(.v15)],
    products: [.library(name: "MyAppComplex", targets: ["MyAppComplex"])],
    targets: [.target(name: "MyAppComplex")]
)
