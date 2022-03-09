function envOrDefault(key: string, defaultValue: string): string {
  return process.env[key] || defaultValue;
}

export default envOrDefault;
