import {ImageUpload} from "./components/ImageUpload.tsx";
import '@mantine/core/styles.css';

import { MantineProvider } from '@mantine/core';

function App() {

  return (
    <MantineProvider defaultColorScheme={'dark'}>
      <ImageUpload />
    </MantineProvider>
  )
}

export default App
