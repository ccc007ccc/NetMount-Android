# Troubleshooting

## Common Issues

### Installation Problems

**Q: KernelSU Manager cannot install module**
- Ensure downloaded file is in `.zip` format
- Check if KernelSU version is up to date
- Try restarting KernelSU Manager app
- Check if device has sufficient storage space

**Q: No effect after installation and reboot**
- Confirm KernelSU is working properly (other modules working)
- Check if `/data/adb/modules/NetMount-Android` directory exists
- Review KernelSU logs for error messages

### Access Problems

**Q: Cannot access Web UI (http://localhost:8088)**
- Wait 2-3 minutes for service to fully start
- Check if device is connected to network
- Try using `http://127.0.0.1:8088`
- Check if firewall or security software is blocking

**Q: Web UI interface blank or fails to load**
- Clear browser cache
- Try different browser
- Check if device time is correct
- Restart module: disable and re-enable in KernelSU Manager

### Mount Problems

**Q: Mount failed with authentication error**
- Verify username and password are correct
- Confirm remote server address format is correct
- For SMB: Confirm share name is correct
- For WebDAV: Confirm URL path is complete

**Q: Mount successful but directory empty**
- Check if network connection is stable
- Confirm remote directory actually contains files
- Wait 5-10 seconds for remote content to load
- Check remote server permission settings

**Q: Failed to mount to /sdcard directory**
- Ensure device is fully unlocked (screen lock password entered)
- Wait for device boot completion before attempting mount
- Check storage permissions are normal

### Performance Issues

**Q: File access very slow**
- Check network connection quality
- Try changing mount parameter configuration
- Consider using wired network instead of WiFi
- Check remote server performance

**Q: Device running sluggishly**
- Reduce number of simultaneously mounted network storages
- Avoid running large file operations on network storage
- Monitor device memory usage

## Network Storage Specific Issues

### SMB/CIFS

**Common errors and solutions:**

- **"Access denied"**: Check user permissions, confirm account has share access rights
- **"Network unreachable"**: Check IP address and network connection
- **"Protocol negotiation failed"**: May be SMB version compatibility issue, contact administrator

**SMB address format:**
```
Correct: 192.168.1.100/sharename
Wrong: \\192.168.1.100\sharename
Wrong: smb://192.168.1.100/sharename
```

### WebDAV

**Common errors and solutions:**

- **SSL certificate error**: 
  - Use `http://` instead of `https://` for testing
  - Or contact administrator to fix certificate issues
- **401 Unauthorized**: Check username/password, confirm account status is normal
- **404 Not Found**: Check if WebDAV path is correct

**WebDAV address format examples:**
```
Nextcloud: https://your-domain.com/remote.php/dav/files/username/
ownCloud: https://your-domain.com/remote.php/webdav/
```

### FTP

**Common errors and solutions:**

- **Connection timeout**: Check firewall settings, confirm FTP ports are open
- **Passive mode issues**: Passive mode is more stable in most cases
- **Anonymous access failed**: Confirm server supports anonymous access

## Log Analysis

### View Real-time Logs

In Web UI "Logs" page you can view real-time running status:

- ✅ **Green info**: Normal operations
- ⚠️ **Yellow warnings**: Non-fatal issues but need attention
- ❌ **Red errors**: Serious problems requiring attention

### Common Log Messages

**Normal startup:**
```
--- NetMount daemon startup ---
Using config file: /data/adb/netmount/config.json
Server listening on :8088...
```

**Successful mount:**
```
✅ Successfully started mount process for 'MyCloud', PID: 12345
✅ Mount verification successful: '/sdcard/NetMount/MyCloud' contains 15 items
```

**Common error messages:**
- `rclone obscure command execution failed`: rclone binary file issue
- `Failed to create mount point directory`: Permission or storage space issue
- `required key not available`: Device encryption not unlocked

## Advanced Troubleshooting

### Manual Service Status Check

Through ADB or terminal emulator:

```bash
# Check if process is running
ps | grep netmountd

# Check port listening
netstat -tlnp | grep 8088

# Check mount points
mount | grep rclone

# View module files
ls -la /data/adb/modules/NetMount-Android/
```

### Reset Configuration

If configuration is corrupted, delete config file to start over:

```bash
# Through ADB
adb shell rm /data/adb/netmount/config.json

# Or in terminal emulator
su
rm /data/adb/netmount/config.json
```

### Complete Module Reinstall

1. Uninstall module in KernelSU Manager
2. Reboot device
3. Reinstall module zip file
4. Reboot device again

## Getting Help

If above methods cannot solve the problem:

1. **Collect Information**:
   - Device model and Android version
   - KernelSU version
   - Module version
   - Detailed error messages and logs

2. **Submit Issue**:
   - GitHub Issues: [project-url]/issues
   - Include complete error logs
   - Describe reproduction steps

3. **Community Support**:
   - Related Android forums
   - KernelSU official groups