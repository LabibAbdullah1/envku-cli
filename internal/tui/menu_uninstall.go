package tui

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/LabibAbdullah1/envku-cli/internal/uninstaller"
)

func runUninstallMenu(reader *bufio.Reader) {
	clearScreen()
	uninst := uninstaller.New()

	fmt.Println("==================================================")
	fmt.Println("             🗑️  UNINSTALL ENVKU CLI             ")
	fmt.Println("==================================================")
	fmt.Println("Pilih salah satu metode pencopotan aplikasi:")
	fmt.Println("")
	fmt.Println("[1] 🛡️  Hapus Envku CLI Saja (DATA & KONFIGURASI AMAN)")
	fmt.Println("    • Menghapus binary 'envku' dan background service Systemd.")
	fmt.Println("    • Database MySQL, file website (/var/www), Nginx, PHP, dan SSL")
	fmt.Println("      TETAP AMAN 100% dan web server tetap berjalan normal.")
	fmt.Println("")
	fmt.Println("[2] 💣 Hapus Bersih Seluruhnya (PURGE / COMPLETE WIPE)")
	fmt.Println("    • Menghapus Envku CLI dan SELURUH paket web stack.")
	fmt.Println("    • Menghapus direktori website (/var/www) dan database (/var/lib/mysql).")
	fmt.Println("    • ⚠️  PERINGATAN: Semua data dan website akan HILANG PERMANEN!")
	fmt.Println("--------------------------------------------------")
	fmt.Println("[0] Batalkan / Kembali ke Menu Utama")
	fmt.Println("==================================================")
	fmt.Print("Pilih opsi [0-2]: ")

	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	switch choice {
	case "1":
		fmt.Print("\nYakin ingin mencopot Envku CLI? (Data web & database tetap aman) [y/N]: ")
		confirm, _ := reader.ReadString('\n')
		if strings.ToLower(strings.TrimSpace(confirm)) == "y" {
			if err := uninst.UninstallSafe(); err != nil {
				fmt.Printf("[-] Error: %v\n", err)
				waitEnter(reader)
			} else {
				fmt.Println("\n==================================================")
				fmt.Println(" [SUKSES] Envku CLI berhasil dicopot!")
				fmt.Println(" Seluruh data web & database Anda tetap aman & aktif.")
				fmt.Println("==================================================")
				os.Exit(0)
			}
		}

	case "2":
		fmt.Println("\n⚠️  PERINGATAN TINGKAT TINGGI!")
		fmt.Println("Tindakan ini akan menghapus Nginx, MySQL, PHP, seluruh file web (/var/www),")
		fmt.Println("dan seluruh database MySQL secara PERMANEN.")
		fmt.Print("\nKetik 'HAPUS' (huruf kapital) untuk konfirmasi: ")
		confirmText, _ := reader.ReadString('\n')
		confirmText = strings.TrimSpace(confirmText)

		if confirmText == "HAPUS" || confirmText == "PURGE" {
			opts := uninstaller.PurgeOptions{
				RemoveWebData:     true,
				RemoveDatabases:   true,
				RemovePackages:    true,
				RemoveHostsConfig: true,
			}
			if err := uninst.UninstallPurge(opts); err != nil {
				fmt.Printf("[-] Error: %v\n", err)
				waitEnter(reader)
			} else {
				fmt.Println("\n==================================================")
				fmt.Println(" [SUKSES] Seluruh sistem telah dibersihkan sepenuhnya.")
				fmt.Println("==================================================")
				os.Exit(0)
			}
		} else {
			fmt.Println("[-] Konfirmasi tidak sesuai. Pencopotan dibatalkan.")
			waitEnter(reader)
		}

	case "0":
		return
	}
}

