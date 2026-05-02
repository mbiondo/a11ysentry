import React from 'react';
import { TouchableOpacity, Text } from 'react-native';

interface Props {
  title: string;
}

const CustomButton = ({ title }: Props) => {
  return (
    <TouchableOpacity
      onPress={() => {}}
      accessibilityLabel={`Submit ${title}`}
      accessibilityRole="button"
    >
      <Text>{title}</Text>
    </TouchableOpacity>
  );
};

export default CustomButton;
