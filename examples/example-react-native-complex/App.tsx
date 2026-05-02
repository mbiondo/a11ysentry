import React from 'react';
import { View, Text, Image, TouchableOpacity, Pressable, ScrollView, StyleSheet } from 'react-native';
import CustomButton from './components/CustomButton';

const App = () => {
  return (
    <ScrollView style={styles.container}>
      <Text style={styles.title}>Complex RN A11y Test</Text>

      {/* Rule: WCAG 1.1.1 (Image without alt/accessibilityLabel) - Error */}
      <Image
        source={{ uri: 'https://reactnative.dev/img/tiny_logo.png' }}
        style={styles.logo}
      />

      {/* Rule: WCAG 1.1.1 (Image with accessibilityLabel) - OK */}
      <Image
        source={{ uri: 'https://reactnative.dev/img/tiny_logo.png' }}
        style={styles.logo}
        accessibilityLabel="React Native Logo"
      />

      <CustomButton title="Save Profile" />

      {/* Rule: WCAG 4.1.2 (Pressable without label) - Error */}
      <Pressable onPress={() => {}} style={styles.button}>
        <View style={styles.icon} />
      </Pressable>

      {/* Rule: WCAG 4.1.2 (TouchableOpacity with label) - OK */}
      <TouchableOpacity 
        onPress={() => {}} 
        accessibilityLabel="Delete Account"
        accessibilityRole="button"
      >
        <Text>Delete</Text>
      </TouchableOpacity>

      {/* Rule: Hidden Focus equivalent (importantForAccessibility="no-hide-descendants") */}
      <View importantForAccessibility="no-hide-descendants">
        <TouchableOpacity onPress={() => {}}>
          <Text>This button is hidden from Screen Readers (Android)</Text>
        </TouchableOpacity>
      </View>

      {/* Rule: aria-hidden equivalent (accessibilityElementsHidden) */}
      <View accessibilityElementsHidden={true}>
        <Pressable onPress={() => {}}>
          <Text>This button is hidden from Screen Readers (iOS)</Text>
        </Pressable>
      </View>
    </ScrollView>
  );
};

const styles = StyleSheet.create({
  container: { padding: 20 },
  title: { fontSize: 24, fontWeight: 'bold' },
  logo: { width: 50, height: 50 },
  button: { width: 44, height: 44, backgroundColor: 'blue' },
  icon: { width: 20, height: 20, backgroundColor: 'white' }
});

export default App;
