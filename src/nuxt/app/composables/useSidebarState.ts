// Delat tillstånd för sidomenyn, så att både ikonen och leader-kommandot
// kan öppna och stänga samma meny.
const isSidebarOpen = useState("sidebar:open", () => false);

export function useSidebarState() {
  function toggleSidebar() {
    isSidebarOpen.value = !isSidebarOpen.value;
  }

  function closeSidebar() {
    isSidebarOpen.value = false;
  }

  return {
    isSidebarOpen,
    toggleSidebar,
    closeSidebar,
  };
}
