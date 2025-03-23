# Dumie

### Dumie, a smart on-demand instance manager

Dumie is a CLI tool designed to help you easily manage dummy instances used for testing purposes. 
It provides an automated and simple command-line interface that helps reduce cloud costs in testing environments. 
By tracking the active status of instances, it automatically terminates them to save costs and automatically saves the work state of testing environments.

### Main Features
1. **Smart Instance Management**  
   Automatically create, update, or delete your instances based on your personalized needs. Dumie saves your work by creating snapshots and restores them when needed. No more worrying about forgetting to shut things down.

2. **Easy Configuration**  
   Configure instance settings via CLI — including firewall rules. Dumie can automatically detect and allowlist your IP address, making it ideal for high-security environments.

3. **Instant Connections**  
   Connect to your instance instantly using the `connect` command. One step, and you're in.


### Dumie has four types of managers:
1. Active Manager: This manager automatically terminates instances when they are not in use.
2. Schedule Manager: This manager automatically terminates instances based on a schedule.
3. TTL Manager: This manager automatically terminates instances after a certain period of time.
4. Manual Manager: This manager allows you to manually manage instances.

### Open for Contributions

I'm open to suggestions and discussions regarding this project. Feel free to open an issue or submit a pull request—I will carefully review your contribution and apply it if it aligns with the project.

For external suggestions, please reach out via chanhyeok.seo2@gmail.com.
