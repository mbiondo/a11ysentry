'use client';

import React from 'react';

export default function Error({ error }: { error: Error }) {
  return (
    <div role="alert" style={{ padding: '20px', border: '1px solid red' }}>
      <h2>Something went wrong!</h2>
      <p>{error.message}</p>
    </div>
  );
}
