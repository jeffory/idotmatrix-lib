package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/jeffory/idotmatrix-lib/device"
	"github.com/jeffory/idotmatrix-lib/protocol"
	"github.com/jeffory/idotmatrix-lib/transport"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	address     string
	displaySize int
	timeout     int
	contrast    float64
	saturation  float64
)

var rootCmd = &cobra.Command{
	Use:   "idotmatrix",
	Short: "Control iDotMatrix pixel displays",
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVarP(&address, "address", "a", "", "device MAC address (or set IDOTMATRIX_ADDRESS)")
	rootCmd.PersistentFlags().IntVarP(&displaySize, "size", "s", 32, "display size: 16, 32, or 64")
	rootCmd.PersistentFlags().IntVar(&timeout, "timeout", 10, "connection timeout in seconds")
	rootCmd.PersistentFlags().Float64Var(&contrast, "contrast", 0, "image contrast adjustment (-100 to 100)")
	rootCmd.PersistentFlags().Float64Var(&saturation, "saturation", 0, "image saturation adjustment (-100 to 100)")

	viper.BindPFlag("device.address", rootCmd.PersistentFlags().Lookup("address"))
	viper.BindPFlag("image.contrast", rootCmd.PersistentFlags().Lookup("contrast"))
	viper.BindPFlag("image.saturation", rootCmd.PersistentFlags().Lookup("saturation"))
}

func configDir() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		return ""
	}
	return filepath.Join(dir, "idotmatrix")
}

func initConfig() {
	dir := configDir()
	if dir == "" {
		return
	}
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(dir)
	viper.ReadInConfig() // ignore error — config file is optional
}

// saveConfig writes the current viper config to disk, creating the
// config directory and file if they don't exist.
func saveConfig() error {
	dir := configDir()
	if dir == "" {
		return fmt.Errorf("cannot determine config directory")
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}
	configFile := filepath.Join(dir, "config.yaml")
	viper.SetConfigFile(configFile)
	return viper.WriteConfig()
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func getDisplaySize() protocol.DisplaySize {
	switch displaySize {
	case 16:
		return protocol.Display16x16
	case 64:
		return protocol.Display64x64
	default:
		return protocol.Display32x32
	}
}

func getAddress() string {
	if address != "" {
		return address
	}
	if env := os.Getenv("IDOTMATRIX_ADDRESS"); env != "" {
		return env
	}
	return viper.GetString("device.address")
}

func connectDevice() (*device.Device, func(), error) {
	addr := getAddress()
	if addr == "" {
		return nil, nil, fmt.Errorf("device address required (use -a, set IDOTMATRIX_ADDRESS, or run 'discover --save')")
	}
	opts := transport.ConnectOptions{
		Timeout:        time.Duration(timeout) * time.Second,
		ChunkDelay:     1 * time.Second,
		UseFlowControl: true,
	}
	conn, err := transport.Connect(addr, opts)
	if err != nil {
		return nil, nil, fmt.Errorf("connect: %w", err)
	}
	dev := device.New(conn, getDisplaySize())
	cleanup := func() { conn.Close() }
	return dev, cleanup, nil
}
