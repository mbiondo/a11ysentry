package com.example.complex.ui;

import android.os.Bundle;
import androidx.appcompat.app.AppCompatActivity;
import android.widget.Button;
import com.example.complex.R;

public class LegacyActivity extends AppCompatActivity {
    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        setContentView(R.layout.activity_legacy);

        Button btn = findViewById(R.id.btn_java);
        // Accessibility Issue: Setting a content description to empty string programmatically
        btn.setContentDescription("");
    }
}
