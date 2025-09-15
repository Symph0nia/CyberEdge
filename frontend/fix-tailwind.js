// Linus-style "No bullshit" Tailwind CSS removal script
// This script removes all Tailwind classes that cause the system to break

const fs = require('fs');
const path = require('path');

// List of Vue files that need fixing
const filesToFix = [
  'src/components/Task/TaskList.vue',
  'src/components/Task/TaskForm.vue',
  'src/components/Utils/PopupNotification.vue',
  'src/components/Utils/StatItem.vue',
  'src/components/Utils/StatusCard.vue',
  'src/components/Tools/HttpRequestTool.vue',
  'src/components/UnderDevelopment.vue',
  'src/components/User/UserManagement.vue',
  'src/components/Utils/ConfirmDialog.vue',
  'src/components/Target/TargetFormContent.vue',
  'src/components/Subdomain/SubdomainScanTable.vue',
  'src/components/Target/DialogModal.vue',
  'src/components/Target/TargetDetail.vue',
  'src/components/RightSidebarMenu.vue',
  'src/components/Subdomain/SubdomainScanDetail.vue',
  'src/components/Subdomain/SubdomainScanResults.vue',
  'src/components/Path/PathScanResults.vue',
  'src/components/Path/PathScanTable.vue',
  'src/components/Port/PortScanDetail.vue',
  'src/components/Port/PortScanResults.vue',
  'src/components/Login/GoogleAuthQRCode.vue',
  'src/components/Path/PathScanDetail.vue',
  'src/components/FooterPage.vue',
  'src/components/HomePage.vue',
  'src/components/Config/SystemConfiguration.vue',
  'src/components/Config/ToolConfiguration.vue',
  'src/App.vue'
];

// Common Tailwind class patterns to remove
const tailwindPatterns = [
  /\bbg-[a-zA-Z0-9-\/]+/g,
  /\btext-[a-zA-Z0-9-\/]+/g,
  /\bborder-[a-zA-Z0-9-\/]+/g,
  /\bp-[a-zA-Z0-9-\/]+/g,
  /\bm-[a-zA-Z0-9-\/]+/g,
  /\bw-[a-zA-Z0-9-\/\[\]]+/g,
  /\bh-[a-zA-Z0-9-\/\[\]]+/g,
  /\bflex[\w-]*/g,
  /\bgrid[\w-]*/g,
  /\bspace-[a-zA-Z0-9-\/]+/g,
  /\brounded[\w-]*/g,
  /\bshadow[\w-]*/g,
  /\bhover:[a-zA-Z0-9-\/]+/g,
  /\bfocus:[a-zA-Z0-9-\/]+/g,
  /\bdark:[a-zA-Z0-9-\/]+/g,
  /\bsm:[a-zA-Z0-9-\/]+/g,
  /\bmd:[a-zA-Z0-9-\/]+/g,
  /\blg:[a-zA-Z0-9-\/]+/g,
  /\bxl:[a-zA-Z0-9-\/]+/g,
  /\bmax-w-[a-zA-Z0-9-\/\[\]]+/g,
  /\bmin-w-[a-zA-Z0-9-\/\[\]]+/g,
  /\bmax-h-[a-zA-Z0-9-\/\[\]]+/g,
  /\bmin-h-[a-zA-Z0-9-\/\[\]]+/g,
  /\boverflw-[a-zA-Z0-9-\/]+/g,
  /\bhidden/g,
  /\bblock/g,
  /\binline[\w-]*/g,
  /\bfixed/g,
  /\babsolute/g,
  /\brelative/g,
  /\bsticky/g,
  /\btop-[a-zA-Z0-9-\/\[\]]+/g,
  /\bbottom-[a-zA-Z0-9-\/\[\]]+/g,
  /\bleft-[a-zA-Z0-9-\/\[\]]+/g,
  /\bright-[a-zA-Z0-9-\/\[\]]+/g,
  /\bz-[a-zA-Z0-9-\/]+/g,
  /\btransition[\w-]*/g,
  /\btransform/g,
  /\bscale-[a-zA-Z0-9-\/]+/g,
  /\brotate-[a-zA-Z0-9-\/]+/g,
  /\btranslate-[a-zA-Z0-9-\/]+/g,
  /\bopacity-[a-zA-Z0-9-\/]+/g,
  /\bbackdrop-[\w-]+/g,
  /\bcontainer/g,
  /\bmx-auto/g,
  /\bpx-[a-zA-Z0-9-\/]+/g,
  /\bpy-[a-zA-Z0-9-\/]+/g,
  /\bpt-[a-zA-Z0-9-\/]+/g,
  /\bpb-[a-zA-Z0-9-\/]+/g,
  /\bpl-[a-zA-Z0-9-\/]+/g,
  /\bpr-[a-zA-Z0-9-\/]+/g,
  /\bmt-[a-zA-Z0-9-\/]+/g,
  /\bmb-[a-zA-Z0-9-\/]+/g,
  /\bml-[a-zA-Z0-9-\/]+/g,
  /\bmr-[a-zA-Z0-9-\/]+/g,
  /\bgap-[a-zA-Z0-9-\/]+/g,
  /\bjustify-[\w-]+/g,
  /\bitems-[\w-]+/g,
  /\bself-[\w-]+/g,
  /\bplace-[\w-]+/g,
  /\bcol-span-[a-zA-Z0-9-\/]+/g,
  /\brow-span-[a-zA-Z0-9-\/]+/g,
  /\bfont-[\w-]+/g,
  /\bleading-[\w-]+/g,
  /\btracking-[\w-]+/g,
  /\bunderline/g,
  /\bline-through/g,
  /\bno-underline/g,
  /\buppercase/g,
  /\blowercase/g,
  /\bcapitalize/g,
  /\bnormal-case/g,
  /\bitalic/g,
  /\bnot-italic/g,
  /\bcursor-[\w-]+/g,
  /\bselect-[\w-]+/g,
  /\bpointer-events-[\w-]+/g,
  /\buser-select-[\w-]+/g,
  /\banimate-[\w-]+/g,
  /\bwhitespace-[\w-]+/g,
  /\btruncate/g,
  /\btext-ellipsis/g,
  /\btext-clip/g,
  /\bbreak-[\w-]+/g,
  /\boverflow-[\w-]+/g,
  /\bscrollbar-[\w-]+/g
];

function cleanTailwindClasses(content) {
  let cleaned = content;

  // Remove Tailwind classes from class attributes
  cleaned = cleaned.replace(/class="([^"]*?)"/g, (match, classes) => {
    let cleanedClasses = classes;

    // Remove each Tailwind pattern
    tailwindPatterns.forEach(pattern => {
      cleanedClasses = cleanedClasses.replace(pattern, '');
    });

    // Clean up extra spaces and trim
    cleanedClasses = cleanedClasses.replace(/\s+/g, ' ').trim();

    // If no classes left, remove the entire class attribute
    if (!cleanedClasses) {
      return '';
    }

    return `class="${cleanedClasses}"`;
  });

  return cleaned;
}

function fixFile(filePath) {
  const fullPath = path.join(__dirname, filePath);

  if (!fs.existsSync(fullPath)) {
    console.log(`⚠️  File not found: ${filePath}`);
    return;
  }

  try {
    const content = fs.readFileSync(fullPath, 'utf8');
    const cleanedContent = cleanTailwindClasses(content);

    if (content !== cleanedContent) {
      fs.writeFileSync(fullPath, cleanedContent, 'utf8');
      console.log(`✅ Fixed: ${filePath}`);
    } else {
      console.log(`✔️  No changes needed: ${filePath}`);
    }
  } catch (error) {
    console.error(`❌ Error fixing ${filePath}:`, error.message);
  }
}

// Main execution
console.log('🚀 Starting Linus-style Tailwind cleanup...\n');

filesToFix.forEach(fixFile);

console.log('\n🎉 Tailwind cleanup completed!');
console.log('💡 System should now work without Tailwind CSS dependencies.');