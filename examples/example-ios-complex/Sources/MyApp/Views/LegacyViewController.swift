import UIKit

class LegacyViewController: UIViewController {
    override func viewDidLoad() {
        super.viewDidLoad()
        
        let img = UIImageView(image: UIImage(named: "logo"))
        // Missing isAccessibilityElement or accessibilityLabel
        view.addSubview(img)
    }
}
