package com.example.complex.components

import androidx.compose.runtime.Composable
import androidx.compose.foundation.Image
import androidx.compose.ui.res.painterResource

@Composable
fun CustomHeader(title: String) {
    // Accessibility Issue: Image without contentDescription in a sub-component
    Image(
        painter = painterResource(id = 1),
        contentDescription = null
    )
}
